package asset_test

import (
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func TestExtractOptimusQueryVersions(t *testing.T) {
	tests := []struct {
		name            string
		data            map[string]interface{}
		wantSQL         int
		wantResolvedSQL int
	}{
		{
			name:            "returns zeros when optimus key is absent",
			data:            map[string]interface{}{},
			wantSQL:         0,
			wantResolvedSQL: 0,
		},
		{
			name:            "returns zeros when optimus value is not a map",
			data:            map[string]interface{}{"optimus": "not-a-map"},
			wantSQL:         0,
			wantResolvedSQL: 0,
		},
		{
			name: "reads int values set in-process",
			data: map[string]interface{}{
				"optimus": map[string]interface{}{
					"sql_version":          3,
					"resolved_sql_version": 2,
				},
			},
			wantSQL:         3,
			wantResolvedSQL: 2,
		},
		{
			name: "reads float64 values as returned by JSON unmarshal",
			data: map[string]interface{}{
				"optimus": map[string]interface{}{
					"sql_version":          float64(5),
					"resolved_sql_version": float64(4),
				},
			},
			wantSQL:         5,
			wantResolvedSQL: 4,
		},
		{
			name: "returns zeros when version keys are missing from optimus map",
			data: map[string]interface{}{
				"optimus": map[string]interface{}{"sql": "SELECT 1"},
			},
			wantSQL:         0,
			wantResolvedSQL: 0,
		},
		{
			name: "returns zero for resolved when only sql_version is present",
			data: map[string]interface{}{
				"optimus": map[string]interface{}{"sql_version": 2},
			},
			wantSQL:         2,
			wantResolvedSQL: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSQL, gotResolved := asset.ExtractOptimusQueryVersions(tc.data)
			assert.Equal(t, tc.wantSQL, gotSQL)
			assert.Equal(t, tc.wantResolvedSQL, gotResolved)
		})
	}
}

func TestInitOptimusQueryVersions(t *testing.T) {
	tests := []struct {
		name                    string
		data                    map[string]interface{}
		wantResolvedInitialized bool
		wantData                map[string]interface{}
	}{
		{
			name:                    "no-op when optimus key is absent",
			data:                    map[string]interface{}{},
			wantResolvedInitialized: false,
			wantData:                map[string]interface{}{},
		},
		{
			name:                    "no-op when optimus value is not a map",
			data:                    map[string]interface{}{"optimus": "not-a-map"},
			wantResolvedInitialized: false,
			wantData:                map[string]interface{}{"optimus": "not-a-map"},
		},
		{
			name: "sets sql_version=1 when sql is present",
			data: map[string]interface{}{
				"optimus": map[string]interface{}{"sql": "SELECT 1"},
			},
			wantResolvedInitialized: false,
			wantData: map[string]interface{}{
				"optimus": map[string]interface{}{
					"sql":         "SELECT 1",
					"sql_version": 1,
				},
			},
		},
		{
			name: "sets both versions and returns true when resolved_sql is present",
			data: map[string]interface{}{
				"optimus": map[string]interface{}{
					"sql":          "SELECT 1",
					"resolved_sql": "SELECT 1 resolved",
				},
			},
			wantResolvedInitialized: true,
			wantData: map[string]interface{}{
				"optimus": map[string]interface{}{
					"sql":                  "SELECT 1",
					"resolved_sql":         "SELECT 1 resolved",
					"sql_version":          1,
					"resolved_sql_version": 1,
				},
			},
		},
		{
			name: "no-op when optimus is empty map",
			data: map[string]interface{}{
				"optimus": map[string]interface{}{},
			},
			wantResolvedInitialized: false,
			wantData: map[string]interface{}{
				"optimus": map[string]interface{}{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := asset.InitOptimusQueryVersions(tc.data)
			assert.Equal(t, tc.wantResolvedInitialized, got)
			assert.Equal(t, tc.wantData, tc.data)
		})
	}
}

func TestBumpOptimusQueryVersions(t *testing.T) {
	changelog := func(path string) diff.Changelog {
		return diff.Changelog{{Type: "update", Path: []string{"data", "optimus", path}, From: "old", To: "new"}}
	}

	tests := []struct {
		name               string
		oldData            map[string]interface{}
		newData            map[string]interface{}
		changelog          diff.Changelog
		wantResolvedInited bool
		wantNewOptimus     map[string]interface{}
	}{
		{
			name:               "no-op when neither sql nor resolved_sql changed",
			oldData:            map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1", "sql_version": 1}},
			newData:            map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1", "sql_version": 1}},
			changelog:          diff.Changelog{},
			wantResolvedInited: false,
			wantNewOptimus:     map[string]interface{}{"sql": "SELECT 1", "sql_version": 1},
		},
		{
			name:               "bumps sql_version when sql changes",
			oldData:            map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1", "sql_version": 1}},
			newData:            map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 2"}},
			changelog:          changelog("sql"),
			wantResolvedInited: false,
			wantNewOptimus:     map[string]interface{}{"sql": "SELECT 2", "sql_version": 2},
		},
		{
			name: "initializes resolved_sql_version to 1 on first appearance",
			oldData: map[string]interface{}{"optimus": map[string]interface{}{
				"sql":         "SELECT 1",
				"sql_version": 1,
			}},
			newData: map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 1",
				"resolved_sql": "SELECT resolved",
			}},
			changelog:          changelog("resolved_sql"),
			wantResolvedInited: true,
			wantNewOptimus: map[string]interface{}{
				"sql":                  "SELECT 1",
				"resolved_sql":         "SELECT resolved",
				"sql_version":          1,
				"resolved_sql_version": 1,
			},
		},
		{
			name: "does not reinitialise resolved_sql_version when it already exists",
			oldData: map[string]interface{}{"optimus": map[string]interface{}{
				"sql":                  "SELECT 1",
				"resolved_sql":         "SELECT resolved",
				"sql_version":          1,
				"resolved_sql_version": 1,
			}},
			newData: map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 1",
				"resolved_sql": "SELECT resolved again",
			}},
			changelog:          changelog("resolved_sql"),
			wantResolvedInited: false,
			wantNewOptimus: map[string]interface{}{
				"sql":                  "SELECT 1",
				"resolved_sql":         "SELECT resolved again",
				"sql_version":          1,
				"resolved_sql_version": 1,
			},
		},
		{
			name: "bumps sql_version and initializes resolved_sql_version together",
			oldData: map[string]interface{}{"optimus": map[string]interface{}{
				"sql":         "SELECT 1",
				"sql_version": 1,
			}},
			newData: map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 2",
				"resolved_sql": "SELECT resolved",
			}},
			changelog: diff.Changelog{
				{Type: "update", Path: []string{"data", "optimus", "sql"}, From: "SELECT 1", To: "SELECT 2"},
				{Type: "update", Path: []string{"data", "optimus", "resolved_sql"}, From: nil, To: "SELECT resolved"},
			},
			wantResolvedInited: true,
			wantNewOptimus: map[string]interface{}{
				"sql":                  "SELECT 2",
				"resolved_sql":         "SELECT resolved",
				"sql_version":          2,
				"resolved_sql_version": 1,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := asset.BumpOptimusQueryVersions(tc.oldData, tc.newData, tc.changelog)
			assert.Equal(t, tc.wantResolvedInited, got)
			assert.Equal(t, tc.wantNewOptimus, tc.newData["optimus"])
		})
	}
}

func TestExtractColumnLineageQuery(t *testing.T) {
	tests := []struct {
		name       string
		changelog  diff.Change
		identifier string
		wantQuery  string
		wantErr    string
	}{
		{
			name:       "returns empty when path does not match identifier",
			changelog:  diff.Change{Path: []string{"data", "optimus", "sql"}, To: "SELECT 1"},
			identifier: "data.optimus.resolved_sql",
			wantQuery:  "",
		},
		{
			name:       "returns query when path matches identifier exactly",
			changelog:  diff.Change{Path: []string{"data", "optimus", "resolved_sql"}, To: "SELECT resolved"},
			identifier: "data.optimus.resolved_sql",
			wantQuery:  "SELECT resolved",
		},
		{
			name:       "returns empty when changelog.To is nil",
			changelog:  diff.Change{Path: []string{"data", "optimus", "resolved_sql"}, To: nil},
			identifier: "data.optimus.resolved_sql",
			wantQuery:  "",
		},
		{
			name:       "returns error when To is not a string for direct match",
			changelog:  diff.Change{Path: []string{"data", "optimus", "resolved_sql"}, To: 42},
			identifier: "data.optimus.resolved_sql",
			wantErr:    "query value is not a string",
		},
		{
			name: "traverses nested map and returns query",
			changelog: diff.Change{
				Path: []string{"data", "optimus"},
				To: map[string]interface{}{
					"nested": map[string]interface{}{
						"query": "SELECT nested",
					},
				},
			},
			identifier: "data.optimus.nested.query",
			wantQuery:  "SELECT nested",
		},
		{
			name: "returns empty when nested key is absent from map",
			changelog: diff.Change{
				Path: []string{"data", "optimus"},
				To:   map[string]interface{}{"other": "value"},
			},
			identifier: "data.optimus.nested.query",
			wantQuery:  "",
		},
		{
			name: "returns error when traversal hits a non-map intermediate value",
			changelog: diff.Change{
				Path: []string{"data", "optimus"},
				To:   "not-a-map",
			},
			identifier: "data.optimus.nested.query",
			wantErr:    "expected map while traversing",
		},
		{
			name: "returns error when resolved nested field is not a string",
			changelog: diff.Change{
				Path: []string{"data", "optimus"},
				To:   map[string]interface{}{"query": 99},
			},
			identifier: "data.optimus.query",
			wantErr:    "resolved field",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := asset.ExtractColumnLineageQuery(tc.changelog, tc.identifier)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantQuery, got)
		})
	}
}
