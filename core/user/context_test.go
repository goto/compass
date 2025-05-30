package user_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/goto/compass/core/user"
)

func TestContext(t *testing.T) {
	t.Run("should return passed user if exist in context", func(t *testing.T) {
		passedUser := user.User{Email: "test@test.com"}
		userCtx := user.NewContext(context.Background(), passedUser)
		actual := user.FromContext(userCtx)
		if !cmp.Equal(passedUser, actual) {
			t.Fatalf("actual is \"%+v\" but expected was \"%+v\"", actual, passedUser)
		}
	})

	t.Run("should return empty user if not exist in context", func(t *testing.T) {
		actual := user.FromContext(context.Background())
		if actual != (user.User{}) {
			t.Fatalf("actual is \"%+v\" but expected was \"%+v\"", actual, "")
		}
	})
}
