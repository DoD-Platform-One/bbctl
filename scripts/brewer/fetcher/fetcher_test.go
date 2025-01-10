package fetcher

import (
	"testing"
)

func TestGitlabRelease_SourceTarballURI(t *testing.T) {
	wantURI := "bananas"
	given := GitlabRelease{
		TagName: "",
		Assets: releaseAssets{
			Sources: []releaseSource{
				{
					Format: "not a tarball",
					URL:    "some url",
				},
				{
					Format: tarGzFormat,
					URL:    wantURI,
				},
			},
		},
	}

	got, err := given.SourceTarballURI()
	if err != nil {
		t.Error(err)
	}

	if got != wantURI {
		t.Errorf("got %s, want %s", got, wantURI)
	}
}
