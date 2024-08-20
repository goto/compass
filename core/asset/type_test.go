package asset_test

import (
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/stretchr/testify/assert"
)

func TestTypeString(t *testing.T) {
	t.Run("built_in_types_should_return_expected_string", func(t *testing.T) {
		for _type, expected := range map[asset.Type]string{
			asset.TypeDashboard:    "dashboard",
			asset.TypeJob:          "job",
			asset.TypeTable:        "table",
			asset.TypeTopic:        "topic",
			asset.TypeFeatureTable: "feature_table",
			asset.TypeApplication:  "application",
			asset.TypeModel:        "model",
			asset.TypeQuery:        "query",
			asset.TypeMetric:       "metric",
		} {
			t.Run((string)(_type), func(t *testing.T) {
				assert.Equal(t, expected, _type.String())
			})
		}
	})
}

func TestTypeIsValid(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		validTypes := []string{
			"table",
			"job",
			"dashboard",
			"topic",
			"feature_table",
			"application",
			"model",
			"query",
			"metric",
		}

		for _, _type := range validTypes {
			t.Run(_type, func(t *testing.T) {
				typeToValidate := asset.Type(_type)
				assert.Truef(t, typeToValidate.IsValid(), "%s should be valid", _type)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		typeToValidate := asset.Type("random")

		assert.Falsef(t, typeToValidate.IsValid(), "%s should be invalid")
	})
}
