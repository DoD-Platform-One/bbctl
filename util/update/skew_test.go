package update_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/update"
)

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

//nolint:thelper // These are not helpers. I want to point to the line in the report functions.
func TestSkew(t *testing.T) {
	tests := map[string]struct {
		prepare func(*testing.T) (update.Version, update.Version)
		report  func(*testing.T, update.Skew)
	}{
		"same version": {
			prepare: func(t *testing.T) (update.Version, update.Version) {
				t.Helper()
				return must(update.NewVersion("1.2.3", time.Now())), must(update.NewVersion("1.2.3", time.Now()))
			},
			report: func(t *testing.T, s update.Skew) {
				assert.True(t, s.IsUpToDate())
				assert.False(t, s.HasMajorUpdate())
				assert.False(t, s.HasMinorUpdate())
				assert.False(t, s.HasPatchUpdate())
			},
		},
		"major update": {
			prepare: func(t *testing.T) (update.Version, update.Version) {
				t.Helper()
				return must(update.NewVersion("1.2.3", time.Now())), must(update.NewVersion("2.0.0", time.Now()))
			},
			report: func(t *testing.T, s update.Skew) {
				assert.True(t, s.HasMajorUpdate())
			},
		},
		"minor update": {
			prepare: func(t *testing.T) (update.Version, update.Version) {
				t.Helper()
				return must(update.NewVersion("1.2.3", time.Now())), must(update.NewVersion("1.3.0", time.Now()))
			},
			report: func(t *testing.T, s update.Skew) {
				assert.False(t, s.HasMajorUpdate())
				assert.True(t, s.HasMinorUpdate())
			},
		},
		"patch update": {
			prepare: func(t *testing.T) (update.Version, update.Version) {
				t.Helper()
				return must(update.NewVersion("1.2.3", time.Now())), must(update.NewVersion("1.2.4", time.Now()))
			},
			report: func(t *testing.T, s update.Skew) {
				assert.False(t, s.HasMajorUpdate())
				assert.False(t, s.HasMinorUpdate())
				assert.True(t, s.HasPatchUpdate())
			},
		},
		"with and without v prefix": {
			prepare: func(t *testing.T) (update.Version, update.Version) {
				t.Helper()
				return must(update.NewVersion("v1.2.3-blah+jimbob", time.Now())), must(update.NewVersion("1.2.3-blah+jimbob", time.Now()))
			},
			report: func(t *testing.T, s update.Skew) {
				assert.True(t, s.IsUpToDate())
				assert.Equal(t, s.LatestVersion().String(), s.CurrentVersion().String())
			},
		},
		"at least a month apart": {
			prepare: func(t *testing.T) (update.Version, update.Version) {
				t.Helper()
				return must(update.NewVersion("1.2.3", time.Now())), must(update.NewVersion("1.2.3", time.Now().AddDate(0, 0, 90)))
			},
			report: func(t *testing.T, s update.Skew) {
				assert.True(t, s.MoreThan(time.Hour*24*30))
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v1, v2 := test.prepare(t)
			skew := update.NewSkew(v1, v2)
			test.report(t, skew)
		})
	}
}
