package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"repo1.dso.mil/big-bang/apps/developer-tools/go-utils/yamler"
)

type PackageVersion struct {
	Version         string
	LatestVersion   string
	UpdateAvailable bool
	SHAsMatch       string
}

// ToMap converts a PackageVersion to a map[string]any
// and does some internal filtering to remove fields that are empty
// or are not set
func (v PackageVersion) ToMap() map[string]any {
	outputMap := map[string]any{}
	if v.Version != "" {
		outputMap["version"] = v.Version
	}
	// Only show the updateAvailable flag if the latestVersion has been retrieved
	if v.LatestVersion != "" {
		outputMap["latestVersion"] = v.LatestVersion
		outputMap["updateAvailable"] = v.UpdateAvailable
	}
	if v.SHAsMatch != "" {
		outputMap["shasMatch"] = v.SHAsMatch
	}
	return outputMap
}

func (v *PackageVersion) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(v.ToMap())
}

func (v *PackageVersion) EncodeJSON() ([]byte, error) {
	return json.Marshal(v.ToMap())
}

func (v *PackageVersion) EncodeText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *PackageVersion) String() string {
	var sb strings.Builder

	if v.Version != "" {
		sb.WriteString(fmt.Sprintf("Version: %s\n", v.Version))
	}
	if v.LatestVersion != "" {
		sb.WriteString(fmt.Sprintf("Latest Version: %s\n", v.LatestVersion))
		sb.WriteString(fmt.Sprintf("Update Available: %t\n", v.UpdateAvailable))
	}
	if v.SHAsMatch != "" {
		sb.WriteString(fmt.Sprintf("SHAs Match: %s\n", v.SHAsMatch))
	}

	return sb.String()
}

type VersionOutput map[string]PackageVersion

func (vo VersionOutput) ToMap() map[string]any {
	outputMap := map[string]any{}
	for chartName, version := range vo {
		outputMap[chartName] = version.ToMap()
	}
	return outputMap
}

func (vo *VersionOutput) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(vo.ToMap())
}

func (vo *VersionOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(vo.ToMap())
}

func (vo *VersionOutput) EncodeText() ([]byte, error) {
	var sb strings.Builder

	for chartName, version := range *vo {
		sb.WriteString(chartName + ":\n")
		sb.WriteString(version.String())
	}

	return []byte(sb.String()), nil
}
