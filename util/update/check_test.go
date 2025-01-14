package update_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/update"
)

func p[T any](t T) *T { return &t }

type errTripper struct {
	http.RoundTripper
}

func (t *errTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("round trip error")
}

//nolint:thelper // These are not helpers. I want to point to the line in the report functions.
func TestGitlabLatestVersion(t *testing.T) {
	tests := map[string]struct {
		prepare func(*testing.T) (context.Context, *http.Client, string)
		report  func(*testing.T, update.Version, error)
	}{
		"nil context": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				return nil, nil, ""
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.Error(t, err)
			},
		},
		"http error": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

				ctx := context.Background()
				client := srv.Client()

				client.Transport = &errTripper{http.DefaultTransport}

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.Error(t, err)
			},
		},
		"invalid json": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte("invalid json"))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.Error(t, err)
			},
		},
		"valid json, but no tags": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.ErrorAs(t, err, p(update.NoTagsError("")))
			},
		},
		"valid json, no commit date": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[{"name":"yooooo"}]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.Error(t, err)
			},
		},
		"valid json, has commit date, not semver": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[{"name":"yooooo","commit":{"created_at":"2021-07-28T00:00:00Z"}}]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.ErrorAs(t, err, p(update.NoValidTagsError("")))
			},
		},
		"valid json, has commit date, valid semver, is prerelease": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[{"name":"v1.2.3-beta","commit":{"created_at":"2021-07-28T00:00:00Z"}}]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.ErrorAs(t, err, p(update.NoValidTagsError("")))
			},
		},
		"valid json, has commit date, valid semver, no release": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[{"name":"v1.2.3","commit":{"created_at":"2021-07-28T00:00:00Z"}}]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.ErrorAs(t, err, p(update.NoValidTagsError("")))
			},
		},
		"valid json, has commit date, valid semver, valid release": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[{"name":"v1.2.3","release":{},"commit":{"created_at":"2021-07-28T00:00:00Z"}}]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL
			},
			report: func(t *testing.T, v update.Version, err error) {
				require.NoError(t, err)
				assert.Equal(t, update.Version{Major: 1, Minor: 2, Patch: 3, ReleasedAt: time.Date(2021, 7, 28, 0, 0, 0, 0, time.UTC)}, v)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, client, url := test.prepare(t)
			version, err := update.GitlabLatestVersion(url)(ctx, client)
			test.report(t, version, err)
		})
	}
}

//nolint:thelper // These are not helpers. I want to point to the line in the report functions.
func TestGitlabSpecificVersion(t *testing.T) {
	tests := map[string]struct {
		prepare func(*testing.T) (context.Context, *http.Client, string, string)
		report  func(*testing.T, update.Version, error)
	}{
		"specific version invalid semver": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string, string) {
				t.Helper()
				return nil, nil, "", "1.2"
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.ErrorAs(t, err, p(update.InvalidSemverError("")))
			},
		},
		"nil context": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string, string) {
				t.Helper()
				return nil, nil, "", "1.2.3"
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.Error(t, err)
			},
		},
		"http error": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

				ctx := context.Background()
				client := srv.Client()

				client.Transport = &errTripper{http.DefaultTransport}

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL, "1.2.3"
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.Error(t, err)
			},
		},
		"invalid json": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte("invalid json"))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL, "1.2.3"
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.Error(t, err)
			},
		},
		"valid json, but tag not found": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL, "1.2.3"
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.ErrorAs(t, err, p(update.NoTagsError("")))
			},
		},
		"valid json, tag found, but not semver": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[{"name":"yooooo"}]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL, "1.2.3"
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.ErrorAs(t, err, p(update.NoValidTagsError("")))
			},
		},
		"valid json, tag found, semver, not up to date": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[{"name":"v1.2.3","commit":{"created_at":"2021-07-28T00:00:00Z"},"release":{}}]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL, "1.2.2"
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.ErrorAs(t, err, p(update.NoValidTagsError("")))
			},
		},
		"valid json, tag found, semver, up to date": {
			prepare: func(t *testing.T) (context.Context, *http.Client, string, string) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte(`[{"name":"v1.2.3","commit":{"created_at":"2021-07-28T00:00:00Z"},"release":{}}]`))
				}))

				ctx := context.Background()
				client := srv.Client()

				t.Cleanup(srv.Close)
				return ctx, client, srv.URL, "1.2.3"
			},
			report: func(t *testing.T, _ update.Version, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, client, url, semver := test.prepare(t)
			version, err := update.GitlabSpecificVersion(url, semver)(ctx, client)
			test.report(t, version, err)
		})
	}
}

//nolint:thelper // These are not helpers. I want to point to the line in the report functions.
func TestCheck(t *testing.T) {
	errFetcher := func(context.Context, *http.Client) (update.Version, error) {
		return update.Version{}, errors.New("fetching version")
	}
	happyFetcher := func(context.Context, *http.Client) (update.Version, error) {
		return update.Version{Major: 1, Minor: 2, Patch: 3, ReleasedAt: time.Now()}, nil
	}
	tests := map[string]struct {
		prepare func(*testing.T) (context.Context, update.VersionFetcher, update.VersionFetcher)
		report  func(*testing.T, update.Skew, error)
	}{
		"error fetching current version": {
			prepare: func(t *testing.T) (context.Context, update.VersionFetcher, update.VersionFetcher) {
				t.Helper()
				return context.Background(), errFetcher, happyFetcher
			},
			report: func(t *testing.T, _ update.Skew, err error) {
				assert.Error(t, err)
			},
		},
		"error fetching latest version": {
			prepare: func(t *testing.T) (context.Context, update.VersionFetcher, update.VersionFetcher) {
				t.Helper()
				return context.Background(), happyFetcher, errFetcher
			},
			report: func(t *testing.T, _ update.Skew, err error) {
				assert.Error(t, err)
			},
		},
		"no error": {
			prepare: func(t *testing.T) (context.Context, update.VersionFetcher, update.VersionFetcher) {
				t.Helper()
				return context.Background(), happyFetcher, happyFetcher
			},
			report: func(t *testing.T, s update.Skew, err error) {
				require.NoError(t, err)
				assert.True(t, s.IsUpToDate())
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cvFetcher, lvFetcher := test.prepare(t)
			skew, err := update.Check(ctx,
				update.WithHTTPClient(http.DefaultClient),
				update.WithCurrentVersionFetcher(cvFetcher),
				update.WithLatestVersionFetcher(lvFetcher),
			)
			test.report(t, skew, err)
		})
	}
}
