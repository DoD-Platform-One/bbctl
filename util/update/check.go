package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/sync/errgroup"
	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
)

func fetchGitLabTags(ctx context.Context, c *http.Client, tagURL string) ([]*gitlab.Tag, error) {
	tags := make([]*gitlab.Tag, 0, 100) // max page size from GitLab API

	resp := new(http.Response)
	resp.Header = make(http.Header)
	// https://docs.gitlab.com/ee/api/rest/#other-pagination-headers
	resp.Header.Add("X-Next-Page", "1")

	for resp.Header.Get("X-Next-Page") != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, tagURL+"?per_page=100&page="+resp.Header.Get("X-Next-Page"), nil)
		if err != nil {
			return nil, fmt.Errorf("creating update check HTTP request: %w", err)
		}

		resp, err = c.Do(req) // resp is reused for the next iteration
		if err != nil {
			return nil, fmt.Errorf("fetching tags from %s: %w", tagURL, err)
		}

		defer resp.Body.Close()

		var pageTags []*gitlab.Tag
		if err := json.NewDecoder(resp.Body).Decode(&pageTags); err != nil {
			return nil, fmt.Errorf("decoding tags from %s: %w", tagURL, err)
		}

		tags = append(tags, pageTags...)
	}

	return tags, nil
}

func tagOk(tag *gitlab.Tag) (Version, bool) {
	if tag.Commit == nil || tag.Commit.CreatedAt == nil {
		return Version{}, false // skip tags without commit date; shouldn't happen, but don't want to panic below
	}

	v, err := NewVersion(tag.Name, *tag.Commit.CreatedAt)
	if err != nil {
		return Version{}, false // skip non-semver tags
	}

	if v.Prerelease != "" && !strings.HasPrefix(v.Prerelease, "bb.") {
		return Version{}, false // skip non-bb. prerelease tags
	}

	if tag.Release == nil {
		return Version{}, false // skip tags without release info
	}

	return v, true
}

// GitlabSpecificVersion returns a VersionFetcher that fetches a specific version
// from a GitLab API [tags endpoint] provided by the tagURL.
//
// [tags endpoint]: https://docs.gitlab.com/ee/api/tags.html
func GitlabSpecificVersion(tagURL string, v string) VersionFetcher {
	return func(ctx context.Context, c *http.Client) (Version, error) {
		sv, err := NewVersion(v, time.Now())
		if err != nil {
			return Version{}, fmt.Errorf("parsing current release from %s: %w", v, err)
		}

		tags, err := fetchGitLabTags(ctx, c, tagURL)
		if err != nil {
			return Version{}, fmt.Errorf("fetching specific release from %s: %w", tagURL, err)
		}

		if len(tags) == 0 {
			return Version{}, NoTagsError("no tags found in " + tagURL)
		}

		for _, tag := range tags {
			if v, ok := tagOk(tag); ok {
				if NewSkew(sv, v).IsUpToDate() {
					return v, nil
				}
			}
		}

		return Version{}, NoValidTagsError("tag " + v + " not found in " + tagURL)
	}
}

// GitlabLatestVersion returns a VersionFetcher that fetches the latest version
// from a GitLab API [tags endpoint] provided by the tagURL.
//
// [tags endpoint]: https://docs.gitlab.com/ee/api/tags.html
func GitlabLatestVersion(tagURL string) VersionFetcher {
	return func(ctx context.Context, c *http.Client) (Version, error) {
		tags, err := fetchGitLabTags(ctx, c, tagURL)
		if err != nil {
			return Version{}, fmt.Errorf("fetching latest release from %s: %w", tagURL, err)
		}

		if len(tags) == 0 {
			return Version{}, NoTagsError("no tags found in " + tagURL)
		}

		for _, tag := range tags {
			if v, ok := tagOk(tag); ok {
				return v, nil
			}
		}

		return Version{}, NoValidTagsError("no valid tags found in " + tagURL)
	}
}

type checkOptions struct {
	currentVersionFetcher VersionFetcher
	latestVersionFetcher  VersionFetcher
	httpClient            *http.Client
}

// CheckOptions allow for overriding the Check operation.
type CheckOption interface {
	apply(co *checkOptions)
}

// A CheckOptionFunc is a function that modifies checkOptions.
// Wrapping with this type allows for passing functions as options.
type CheckOptionFunc func(*checkOptions)

func (f CheckOptionFunc) apply(o *checkOptions) {
	f(o)
}

// WithLatestVersionFetcher sets the function used to fetch the latest version.
func WithLatestVersionFetcher(fetcher VersionFetcher) CheckOption {
	return CheckOptionFunc(func(o *checkOptions) {
		o.latestVersionFetcher = fetcher
	})
}

// WithCurrentVersionFetcher sets the function used to fetch the current version.
func WithCurrentVersionFetcher(fetcher VersionFetcher) CheckOption {
	return CheckOptionFunc(func(o *checkOptions) {
		o.currentVersionFetcher = fetcher
	})
}

// WithHTTPClient sets the HTTP client used to fetch the latest version.
func WithHTTPClient(httpClient *http.Client) CheckOption {
	return CheckOptionFunc(func(o *checkOptions) {
		o.httpClient = httpClient
	})
}

func defaultOptions() *checkOptions {
	constants, _ := static.GetDefaultConstants() // this failing does not affect the logic

	currentVersion := constants.BigBangCliVersion
	if currentVersion == "" {
		currentVersion = "0.0.0"
	}

	const repo1bbctlTagURL = "https://repo1.dso.mil/api/v4/projects/11320/repository/tags"

	return &checkOptions{
		currentVersionFetcher: GitlabSpecificVersion(repo1bbctlTagURL, currentVersion),
		latestVersionFetcher:  GitlabLatestVersion(repo1bbctlTagURL),
		httpClient:            http.DefaultClient,
	}
}

// Check fetches the current version of bbctl and the latest version available
// and returns the skew between them.
func Check(ctx context.Context, opts ...CheckOption) (Skew, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	eg, ctx := errgroup.WithContext(ctx)

	var current Version
	eg.Go(func() error {
		v, err := options.currentVersionFetcher(ctx, options.httpClient)
		if err != nil {
			return fmt.Errorf("fetching current version: %w", err)
		}
		current = v
		return nil
	})

	var latest Version
	eg.Go(func() error {
		v, err := options.latestVersionFetcher(ctx, options.httpClient)
		if err != nil {
			return fmt.Errorf("fetching latest version: %w", err)
		}
		latest = v
		return nil
	})

	if err := eg.Wait(); err != nil {
		return Skew{}, err
	}

	return NewSkew(current, latest), nil
}
