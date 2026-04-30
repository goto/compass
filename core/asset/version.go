package asset

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

const BaseVersion = "0.1"

// ParseVersion returns error if version string is not in MAJOR.MINOR format
func ParseVersion(v string) (*semver.Version, error) {
	semverVersion, err := semver.NewVersion(v)
	if err != nil {
		return nil, fmt.Errorf("invalid version \"%s\"", v)
	}
	return semverVersion, nil
}

// IncreaseMinorVersion bumps up the minor version +0.1.
// If v is empty it is treated as "0.0", yielding "0.1".
func IncreaseMinorVersion(v string) (string, error) {
	if v == "" {
		v = "0.0"
	}
	oldVersion, err := ParseVersion(v)
	if err != nil {
		return "", err
	}
	newVersion := oldVersion.IncMinor()
	return fmt.Sprintf("%d.%d", newVersion.Major(), newVersion.Minor()), nil
}

// IsGreaterThan returns true when new version is strictly greater than old version.
// Invalid or empty versions are treated as "0.0".
func IsGreaterThan(newVersion, oldVersion string) bool {
	parse := func(v string) *semver.Version {
		if v == "" {
			v = "0.0"
		}
		sv, err := semver.NewVersion(v)
		if err != nil {
			sv, _ = semver.NewVersion("0.0")
		}
		return sv
	}
	return parse(newVersion).GreaterThan(parse(oldVersion))
}
