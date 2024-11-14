package user_test

import (
	"testing"

	"github.com/goto/compass/core/user"
)

func TestErrors(t *testing.T) {
	type testCase struct {
		Description    string
		Err            error
		ExpectedString string
	}

	testCases := []testCase{
		{
			Description:    "not found error return correct error string",
			Err:            user.NotFoundError{Email: "test@test.com"},
			ExpectedString: "could not find user with email \"test@test.com\"",
		},
		{
			Description:    "duplicate error return correct error string",
			Err:            user.DuplicateRecordError{Email: "test@test.com"},
			ExpectedString: "duplicate user with email \"test@test.com\"",
		},
		{
			Description:    "invalid error return correct error string",
			Err:            user.InvalidError{Email: "test@test.com"},
			ExpectedString: "empty field with email \"test@test.com\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			if tc.ExpectedString != tc.Err.Error() {
				t.Fatalf("actual is \"%+v\" but expected was \"%+v\"", tc.Err.Error(), tc.ExpectedString)
			}
		})
	}
}
