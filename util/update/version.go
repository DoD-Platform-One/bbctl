package update

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Version represents a semantic version.
type Version struct {
	Major, Minor, Patch int
	Prerelease          string
	BuildMetadata       string
	ReleasedAt          time.Time
}

// String returns a canonical string representation of the version.
func (v Version) String() string {
	sb := new(strings.Builder)

	_, _ = fmt.Fprintf(sb, "%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.Prerelease != "" {
		sb.WriteString("-")
		sb.WriteString(v.Prerelease)
	}

	if v.BuildMetadata != "" {
		sb.WriteString("+")
		sb.WriteString(v.BuildMetadata)
	}

	return sb.String()
}

// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
// We prepend `v?` to the documented regex to allow for a leading `v` in the version string
var semverRGX = regexp.MustCompile(`^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// NewVersion parses a semver string and a time.Time into a Version.
// If the string is not a valid semver string, an error is returned.
func NewVersion(semver string, t time.Time) (Version, error) {
	matches := semverRGX.FindStringSubmatch(semver)
	if len(matches) == 0 {
		return Version{}, InvalidSemverError(semver)
	}

	atoi := func(s string) int {
		i, _ := strconv.Atoi(s) // Regex guarantees this is a number
		return i
	}

	var v Version
	for i, name := range semverRGX.SubexpNames() {
		switch name {
		case "major":
			v.Major = atoi(matches[i])
		case "minor":
			v.Minor = atoi(matches[i])
		case "patch":
			v.Patch = atoi(matches[i])
		case "prerelease":
			v.Prerelease = matches[i]
		case "buildmetadata":
			v.BuildMetadata = matches[i]
		}
	}

	v.ReleasedAt = t

	return v, nil
}

// A VersionFetcher fetches a version of a product. The logic for
// fetching from a particular source must be fully encapsulated in the
// fetcher for that source.
type VersionFetcher func(context.Context, *http.Client) (Version, error)
