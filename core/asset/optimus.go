package asset

import (
	"fmt"
	"strings"

	"github.com/r3labs/diff/v2"
)

const (
	OptimusDataKey               = "optimus"
	OptimusSQLKey                = "sql"
	OptimusResolvedSQLKey        = "resolved_sql"
	OptimusSQLVersionKey         = "sql_version"
	OptimusResolvedSQLVersionKey = "resolved_sql_version"
	OptimusSQLBaseVersion        = 1

	ChangelogPathOptimusSQL         = "data.optimus.sql"
	ChangelogPathOptimusResolvedSQL = "data.optimus.resolved_sql"
)

// ExtractOptimusQueryVersions reads the sql_version and resolved_sql_version
// counters from the asset's data.optimus map.
func ExtractOptimusQueryVersions(data map[string]interface{}) (sqlVersion, resolvedSQLVersion int) {
	optimus, ok := data[OptimusDataKey].(map[string]interface{})
	if !ok {
		return 0, 0
	}
	sqlVersion = toIntVersion(optimus[OptimusSQLVersionKey])
	resolvedSQLVersion = toIntVersion(optimus[OptimusResolvedSQLVersionKey])
	return sqlVersion, resolvedSQLVersion
}

// InitOptimusQueryVersions sets initial version counters when an asset with
// Optimus SQL fields is first inserted. Returns true if resolved_sql was found.
func InitOptimusQueryVersions(data map[string]interface{}) (resolvedSQLInitialized bool) {
	optimus, ok := data[OptimusDataKey].(map[string]interface{})
	if !ok || optimus == nil {
		return false
	}

	if _, found := optimus[OptimusSQLKey].(string); found {
		optimus[OptimusSQLVersionKey] = OptimusSQLBaseVersion
	}

	if _, found := optimus[OptimusResolvedSQLKey].(string); found {
		optimus[OptimusResolvedSQLVersionKey] = OptimusSQLBaseVersion
		resolvedSQLInitialized = true
	}

	return resolvedSQLInitialized
}

// BumpOptimusQueryVersions increments the relevant version counters when SQL
// fields change on an existing asset.  Returns true if resolved_sql_version
// was initialized for the first time (which signals that a lineage fetch is
// required).
func BumpOptimusQueryVersions(oldData, newData map[string]interface{}, changelog diff.Changelog) bool {
	oldOptimus, _ := oldData[OptimusDataKey].(map[string]interface{})
	newOptimus, _ := newData[OptimusDataKey].(map[string]interface{})

	baseChanged := isOptimusFieldChanged(oldOptimus, newOptimus, OptimusSQLKey, changelog, ChangelogPathOptimusSQL)
	resolvedChanged := isOptimusFieldChanged(oldOptimus, newOptimus, OptimusResolvedSQLKey, changelog, ChangelogPathOptimusResolvedSQL)

	if !baseChanged && !resolvedChanged {
		return false
	}

	if newOptimus == nil {
		newOptimus = make(map[string]interface{})
		newData[OptimusDataKey] = newOptimus
	}

	preserveOptimusVersions(oldOptimus, newOptimus)
	return applyOptimusVersionBumps(oldOptimus, newOptimus, baseChanged, resolvedChanged)
}

// ExtractColumnLineageQuery extracts the SQL query string from a changelog To
// value, using columnLineageIdentifier as the target field path.
func ExtractColumnLineageQuery(changelog diff.Change, columnLineageIdentifier string) (string, error) {
	var (
		changelogPath = strings.Join(changelog.Path, ".")
		nestedKey     string
	)

	if strings.HasPrefix(columnLineageIdentifier, changelogPath+".") {
		nestedKey = strings.TrimPrefix(columnLineageIdentifier, changelogPath+".")
	} else if changelogPath != columnLineageIdentifier {
		return "", nil
	}

	if changelog.To == nil {
		return "", nil
	}

	if nestedKey == "" {
		query, ok := changelog.To.(string)
		if !ok {
			return "", fmt.Errorf("query value is not a string")
		}
		return query, nil
	}

	traversedChange := changelog.To
	for _, key := range strings.Split(nestedKey, ".") {
		change, ok := traversedChange.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("expected map while traversing %q", key)
		}
		traversedChange, ok = change[key]
		if !ok {
			return "", nil
		}
	}

	query, ok := traversedChange.(string)
	if !ok {
		return "", fmt.Errorf("resolved field %q is not a string", nestedKey)
	}

	return query, nil
}

func preserveOptimusVersions(oldOptimus, newOptimus map[string]interface{}) {
	for _, key := range []string{OptimusSQLVersionKey, OptimusResolvedSQLVersionKey} {
		if existing := toIntVersion(oldOptimus[key]); existing != 0 {
			if _, alreadySet := newOptimus[key]; !alreadySet {
				newOptimus[key] = existing
			}
		}
	}
}

func applyOptimusVersionBumps(oldOptimus, newOptimus map[string]interface{}, baseChanged, resolvedChanged bool) bool {
	if baseChanged {
		newOptimus[OptimusSQLVersionKey] = nextOptimusVersion(oldOptimus, OptimusSQLVersionKey)
	}

	if resolvedChanged {
		if toIntVersion(oldOptimus[OptimusResolvedSQLVersionKey]) == 0 {
			newOptimus[OptimusResolvedSQLVersionKey] = OptimusSQLBaseVersion
			return true
		}
	}

	return false
}

func nextOptimusVersion(optimus map[string]interface{}, key string) int {
	current := toIntVersion(optimus[key])
	return current + 1
}

func isOptimusFieldChanged(oldOptimus, newOptimus map[string]interface{}, field string, changelog diff.Changelog, changelogPath string) bool {
	for _, c := range changelog {
		if strings.Join(c.Path, ".") == changelogPath {
			return true
		}
	}
	oldVal, _ := oldOptimus[field].(string)
	newVal, _ := newOptimus[field].(string)
	return oldVal != newVal
}

func toIntVersion(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	default:
		return 0
	}
}
