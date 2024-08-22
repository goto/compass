package asset_test

import (
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		assert.True(t, testutils.AreSlicesEqualIgnoringOrder(expectedTypes, actualTypes, compareTypes))
	})
}

func TestRegisterSupportedTypes(t *testing.T) {
	t.Run("should return error if type is invalid", func(t *testing.T) {
		testCases := []struct {
			name             string
			input            asset.Type
			expectedErrorMsg string
		}{
			{
				name:             "empty type",
				input:            asset.Type(""),
				expectedErrorMsg: "type length must be 3 to 16 inclusive",
			},
			{
				name:             "length less than 3",
				input:            asset.Type("ab"),
				expectedErrorMsg: "type length must be 3 to 16 inclusive",
			},
			{
				name:             "length more than 16",
				input:            asset.Type("abcdefghijklmnopq"),
				expectedErrorMsg: "type length must be 3 to 16 inclusive",
			},
			{
				name:             "contains character outside lower case alphanumeric and underscore, with exclamation mark",
				input:            asset.Type("abcd_efgh!"),
				expectedErrorMsg: "type must be combination of alphanumeric and underscores",
			},
			{
				name:             "contains character outside lower case alphanumeric and underscore, with upper case A",
				input:            asset.Type("Abcd_efgh"),
				expectedErrorMsg: "type must be combination of alphanumeric and underscores",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				actualError := asset.RegisterSupportedTypes(tc.input)

				assert.EqualError(t, actualError, tc.expectedErrorMsg)
			})
		}
	})

	t.Run("should not update supported types if one or more types are invalid", func(t *testing.T) {
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

		inputTypes := []asset.Type{
			asset.Type("dataset"),
			asset.Type("invalid!"),
		}

		actualError := asset.RegisterSupportedTypes(inputTypes...)
		require.Error(t, actualError)

		actualTypes := asset.GetSupportedTypes()
		assert.True(t, testutils.AreSlicesEqualIgnoringOrder(expectedTypes, actualTypes, compareTypes))
	})

	t.Run("should update supported types if no error is returned", func(t *testing.T) {
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
			asset.Type("fact_source"),
			asset.Type("dimension_01a"),
		}

		inputTypes := []asset.Type{
			asset.Type("fact_source"),
			asset.Type("dimension_01a"),
		}

		actualError := asset.RegisterSupportedTypes(inputTypes...)
		require.NoError(t, actualError)

		actualTypes := asset.GetSupportedTypes()
		assert.True(t, testutils.AreSlicesEqualIgnoringOrder(expectedTypes, actualTypes, compareTypes))
	})
}

func compareTypes(left, right asset.Type) int {
	if left < right {
		return -1
	}
	if right > left {
		return 1
	}
	return 0
}
