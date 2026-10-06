package asset

import (
	"strings"

	"github.com/r3labs/diff/v2"
)

// ParseExcludedChangelogPathSegments splits dotted exclusion paths for diff filtering.
func ParseExcludedChangelogPathSegments(excludedPaths []string) [][]string {
	if len(excludedPaths) == 0 {
		return nil
	}

	segments := make([][]string, len(excludedPaths))
	for i, path := range excludedPaths {
		segments[i] = strings.Split(path, ".")
	}

	return segments
}

func changelogPathMatchesExcluded(changePath, excludedSegments []string) bool {
	changeIdx := 0
	for excludedIdx := range excludedSegments {
		changeIdx = skipExcludedChangelogPathPrefix(changePath, changeIdx, excludedSegments, excludedIdx)
		if changeIdx >= len(changePath) || changePath[changeIdx] != excludedSegments[excludedIdx] {
			return false
		}
		changeIdx++
	}

	return true
}

func skipExcludedChangelogPathPrefix(changePath []string, changeIdx int, excludedSegments []string, excludedIdx int) int {
	for changeIdx < len(changePath) && isSkippableExcludedChangelogSegment(changePath, changeIdx, excludedSegments, excludedIdx) {
		changeIdx++
	}

	return changeIdx
}

func isSkippableExcludedChangelogSegment(
	changePath []string,
	changeIdx int,
	excludedSegments []string,
	excludedIdx int,
) bool {
	if isChangelogArrayIndex(changePath[changeIdx]) {
		return true
	}

	return excludedIdx > 0 && changePath[changeIdx] == excludedSegments[excludedIdx-1]
}

func changelogChangeExcluded(changePath []string, excludedSegments [][]string) bool {
	for _, segments := range excludedSegments {
		if changelogPathMatchesExcluded(changePath, segments) {
			return true
		}
	}

	return false
}

func isChangelogArrayIndex(segment string) bool {
	if segment == "" {
		return false
	}
	for _, r := range segment {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func filterExcludedChangelog(changelog diff.Changelog, excludedPathSegments [][]string) diff.Changelog {
	if len(excludedPathSegments) == 0 || len(changelog) == 0 {
		return changelog
	}

	var filtered diff.Changelog
	for _, change := range changelog {
		if changelogChangeExcluded(change.Path, excludedPathSegments) {
			continue
		}
		filtered = append(filtered, change)
	}

	if len(filtered) == len(changelog) {
		return changelog
	}

	if len(filtered) == 0 {
		return nil
	}

	return filtered
}
