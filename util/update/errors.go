package update

// NoTagsError is returned from a VersionFetcher when no tags are found.
type NoTagsError string

func (e NoTagsError) Error() string {
	return string(e)
}

// NoValidTagsError is returned from a VersionFetcher when no
// valid releases are found with a GA tag.
type NoValidTagsError string

func (e NoValidTagsError) Error() string {
	return string(e)
}

// InvalidSemverError is returned when a semver string is invalid.
type InvalidSemverError string

func (e InvalidSemverError) Error() string {
	return "invalid semver: " + string(e)
}

// Value returns the invalid semver string.
func (e InvalidSemverError) Value() string {
	return string(e)
}
