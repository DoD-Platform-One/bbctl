package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const repo1Domain = "repo1.dso.mil"
const tarGzFormat = "tar.gz"

type ReleaseInfo struct {
	ReleaseTag string `json:"release_tag"`
	ReleaseURI string `json:"release_uri"`
	Sha256Sum  string `json:"sha256_sum"`
}

type releaseSource struct {
	Format string `json:"format"`
	URL    string `json:"url"`
}

type releaseAssets struct {
	Sources []releaseSource `json:"sources"`
}

type GitlabRelease struct {
	TagName string        `json:"tag_name"`
	Assets  releaseAssets `json:"assets"`
}

type ReleaseFetcher struct {
	ProjectID  int
	urlParseFn func(rawURL string) (*url.URL, error)
	httpClient *http.Client
}

func NewReleaseFetcher(projectID int) *ReleaseFetcher {
	return &ReleaseFetcher{
		ProjectID:  projectID,
		urlParseFn: url.Parse,
		httpClient: &http.Client{},
	}
}

func (glr *GitlabRelease) SourceTarballURI() (string, error) {
	for _, src := range glr.Assets.Sources {
		if src.Format == tarGzFormat {
			return src.URL, nil
		}
	}
	return "", errors.New("could not find source uri for gitlab release")
}

func (rf *ReleaseFetcher) FetchRepo1Uri(uri string) ([]byte, error) {
	parsed, err := rf.urlParseFn(uri)
	if err != nil {
		return nil, errors.New("could not parse url")
	}
	if !strings.Contains(parsed.Hostname(), repo1Domain) {
		return nil, fmt.Errorf("for security reasons I only fetch URLs from %s", repo1Domain)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build a request for %s: %w", uri, err)
	}
	resp, err2 := rf.httpClient.Do(req)
	if err2 != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", uri, err2)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response bytes: %w", err)
	}
	return respBytes, nil
}

func (rf *ReleaseFetcher) FetchLatestReleaseInfo() (GitlabRelease, error) {
	var out GitlabRelease
	releaseURL := fmt.Sprintf("https://repo1.dso.mil/api/v4/projects/%d/releases?per_page=1", rf.ProjectID)
	respBytes, err := rf.FetchRepo1Uri(releaseURL)
	if err != nil {
		return out, err
	}

	releases := make([]GitlabRelease, 0)
	err = json.Unmarshal(respBytes, &releases)
	if err != nil {
		return out, err
	}

	for _, release := range releases {
		return release, nil
	}

	return out, errors.New("could not find latest release")
}
