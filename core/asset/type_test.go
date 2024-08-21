package asset_test

import (
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/stretchr/testify/assert"
)

func TestTypeString(t *testing.T) {
	t.Run("should_return_same_string", func(t *testing.T) {
		testCases := []struct {
			input    asset.Type
			expected string
		}{
			{
				input:    asset.Type("dashboard"),
				expected: "dashboard",
			},
			{
				input:    asset.Type("job"),
				expected: "job",
			},
			{
				input:    asset.Type("table"),
				expected: "table",
			},
			{
				input:    asset.Type("topic"),
				expected: "topic",
			},
			{
				input:    asset.Type("feature_table"),
				expected: "feature_table",
			},
			{
				input:    asset.Type("application"),
				expected: "application",
			},
			{
				input:    asset.Type("model"),
				expected: "model",
			},
			{
				input:    asset.Type("query"),
				expected: "query",
			},
			{
				input:    asset.Type("metric"),
				expected: "metric",
			},
		}

		for _, tc := range testCases {
			actual := tc.input.String()

			assert.Equal(t, tc.expected, actual)
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

func TestGetSupportedTypes(t *testing.T) {
	t.Run("should_return_valid_built_in_types", func(t *testing.T) {
		expectedTypes := []asset.Type{
			asset.Type("table"),
			asset.Type("job"),
			asset.Type("dashboard"),
			asset.Type("topic"),
			asset.Type("feature_table"),
			asset.Type("application"),
			asset.Type("model"),
			asset.Type("query"),
			asset.Type("metric"),
		}

		actualTypes := asset.GetSupportedTypes()

		assert.EqualValues(t, expectedTypes, actualTypes)
	})
}
