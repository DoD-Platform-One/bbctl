package update

import "time"

// Skew represents the difference between two versions.
type Skew struct {
	current Version
	latest  Version
	Major   int
	Minor   int
	Patch   int
	Time    time.Duration
}

// IsUpToDate returns true if the current version is the same as the latest version.
func (s Skew) IsUpToDate() bool {
	return s.Major == 0 && s.Minor == 0 && s.Patch == 0
}

// HasMajorUpdate returns true if the latest version has a higher major version than the current version.
func (s Skew) HasMajorUpdate() bool {
	return s.Major > 0
}

// HasMinorUpdate returns true if the latest version has a higher minor version than the current version.
func (s Skew) HasMinorUpdate() bool {
	return s.Minor > 0
}

// HasPatchUpdate returns true if the latest version has a higher patch version than the current version.
func (s Skew) HasPatchUpdate() bool {
	return s.Patch > 0
}

// MoreThan returns true if the time difference between versions is greater than the given duration.
func (s Skew) MoreThan(dur time.Duration) bool {
	return s.Time > dur
}

// LatestVersion returns the latest version.
func (s Skew) LatestVersion() Version {
	return s.latest
}

// CurrentVersion returns the current version.
func (s Skew) CurrentVersion() Version {
	return s.current
}

// NewSkew calculates the drift between two versions.
func NewSkew(v1 Version, v2 Version) Skew {
	return Skew{
		current: v1,
		latest:  v2,
		Major:   v2.Major - v1.Major,
		Minor:   v2.Minor - v1.Minor,
		Patch:   v2.Patch - v1.Patch,
		Time:    v2.ReleasedAt.Sub(v1.ReleasedAt),
	}
}
