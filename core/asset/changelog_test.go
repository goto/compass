package asset_test

import (
	"encoding/json"
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func excludedSegments(paths ...string) [][]string {
	return asset.ParseExcludedChangelogPathSegments(paths)
}

func TestExcludedChangelogPathsInsideArrays(t *testing.T) {
	source := `{
		"data": {
			"columns": [
				{
					"name": "jakarta_data_date",
					"description": "old description"
				}
			]
		}
	}`
	target := `{
		"data": {
			"columns": [
				{
					"name": "jakarta_data_date",
					"description": "new description"
				}
			]
		}
	}`

	var sourceAsset, targetAsset asset.Asset
	require.NoError(t, json.Unmarshal([]byte(source), &sourceAsset))
	require.NoError(t, json.Unmarshal([]byte(target), &targetAsset))

	full, simplified, err := sourceAsset.Diff(&targetAsset, excludedSegments("data.columns.description"))
	require.NoError(t, err)
	require.Len(t, full, 1)
	assert.Equal(t, []string{"data", "columns", "0", "description"}, full[0].Path)
	assert.Empty(t, simplified)
}

func TestExcludedChangelogPathsNestedColumnArrays(t *testing.T) {
	source := `{
		"data": {
			"columns": [
				{
					"name": "parent",
					"columns": [
						{
							"name": "child",
							"description": "old"
						}
					]
				}
			]
		}
	}`
	target := `{
		"data": {
			"columns": [
				{
					"name": "parent",
					"columns": [
						{
							"name": "child",
							"description": "new"
						}
					]
				}
			]
		}
	}`

	var sourceAsset, targetAsset asset.Asset
	require.NoError(t, json.Unmarshal([]byte(source), &sourceAsset))
	require.NoError(t, json.Unmarshal([]byte(target), &targetAsset))

	_, simplified, err := sourceAsset.Diff(&targetAsset, excludedSegments("data.columns.description"))
	require.NoError(t, err)
	assert.Empty(t, simplified)
}

func TestExcludedChangelogPathsMultipleTopLevelColumns(t *testing.T) {
	source := `{
		"data": {
			"columns": [
				{"name": "a", "description": "old-a"},
				{"name": "b", "description": "old-b"}
			]
		}
	}`
	target := `{
		"data": {
			"columns": [
				{"name": "a", "description": "new-a"},
				{"name": "b", "description": "new-b"}
			]
		}
	}`

	var sourceAsset, targetAsset asset.Asset
	require.NoError(t, json.Unmarshal([]byte(source), &sourceAsset))
	require.NoError(t, json.Unmarshal([]byte(target), &targetAsset))

	full, simplified, err := sourceAsset.Diff(&targetAsset, excludedSegments("data.columns.description"))
	require.NoError(t, err)
	require.Len(t, full, 2)
	assert.Empty(t, simplified)
}

func TestExcludedChangelogPathsMixedExcludedAndTrackedColumnChanges(t *testing.T) {
	source := `{
		"data": {
			"columns": [
				{"name": "a", "description": "old-a"},
				{"name": "old-b", "description": "old-llm-b"}
			]
		}
	}`
	target := `{
		"data": {
			"columns": [
				{"name": "a", "description": "new-a"},
				{"name": "new-b", "description": "new-llm-b"}
			]
		}
	}`

	var sourceAsset, targetAsset asset.Asset
	require.NoError(t, json.Unmarshal([]byte(source), &sourceAsset))
	require.NoError(t, json.Unmarshal([]byte(target), &targetAsset))

	_, simplified, err := sourceAsset.Diff(&targetAsset, excludedSegments("data.columns.description"))
	require.NoError(t, err)
	require.Len(t, simplified, 1)
	assert.Equal(t, []string{"data", "columns", "1", "name"}, simplified[0].Path)
}

func TestExcludedChangelogPathsKeepsOtherColumnFields(t *testing.T) {
	source := `{
		"data": {
			"columns": [
				{
					"name": "old_name",
					"description": "same"
				}
			]
		}
	}`
	target := `{
		"data": {
			"columns": [
				{
					"name": "new_name",
					"description": "same"
				}
			]
		}
	}`

	var sourceAsset, targetAsset asset.Asset
	require.NoError(t, json.Unmarshal([]byte(source), &sourceAsset))
	require.NoError(t, json.Unmarshal([]byte(target), &targetAsset))

	_, simplified, err := sourceAsset.Diff(&targetAsset, excludedSegments("data.columns.description"))
	require.NoError(t, err)
	require.Len(t, simplified, 1)
	assert.Equal(t, []string{"data", "columns", "0", "name"}, simplified[0].Path)
}

func TestFilterExcludedChangelogReusesChangelogWhenNothingRemoved(t *testing.T) {
	source := `{"data": {"title": "old"}}`
	target := `{"data": {"title": "new"}}`

	var sourceAsset, targetAsset asset.Asset
	require.NoError(t, json.Unmarshal([]byte(source), &sourceAsset))
	require.NoError(t, json.Unmarshal([]byte(target), &targetAsset))

	full, simplified, err := sourceAsset.Diff(&targetAsset, excludedSegments("data.columns.description"))
	require.NoError(t, err)
	require.NotEmpty(t, full)
	assert.Equal(t, full, simplified)
}

func TestExcludedChangelogPathsStillFiltersTopLevelDataFields(t *testing.T) {
	source := `{"data": {"update_time": "2024-01-01T00:00:00Z", "title": "old"}}`
	target := `{"data": {"update_time": "2024-01-02T00:00:00Z", "title": "new"}}`

	var sourceAsset, targetAsset asset.Asset
	require.NoError(t, json.Unmarshal([]byte(source), &sourceAsset))
	require.NoError(t, json.Unmarshal([]byte(target), &targetAsset))

	full, simplified, err := sourceAsset.Diff(&targetAsset, excludedSegments("data.update_time"))
	require.NoError(t, err)
	require.Len(t, full, 2)
	require.Len(t, simplified, 1)
	assert.Equal(t, []string{"data", "title"}, simplified[0].Path)
}
