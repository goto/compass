package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/goto/compass/core/user"
	"github.com/goto/compass/core/user/mocks"
	"github.com/goto/salt/log"
	"github.com/stretchr/testify/assert"
)

func TestValidateUser(t *testing.T) {
	userID := "user-id"
	type testCase struct {
		Description string
		Email       string
		Setup       func(ctx context.Context, inputEmail string, userRepo *mocks.UserRepository)
		ExpectErr   error
	}

	testCases := []testCase{
		{
			Description: "should return no user error when email is empty",
			ExpectErr:   user.ErrNoUserInformation,
		},
		{
			Description: "should return user ID when successfully found user Email from DB, or not found but can create the new one",
			Email:       "test@test.com",
			Setup: func(ctx context.Context, inputEmail string, userRepo *mocks.UserRepository) {
				userRepo.EXPECT().GetOrInsertByEmail(ctx, &user.User{Email: inputEmail}).Return(userID, nil)
			},
			ExpectErr: nil,
		},
		{
			Description: "should return user error if InsertByEmail return empty ID and error",
			Email:       "test@test.com",
			Setup: func(ctx context.Context, inputEmail string, userRepo *mocks.UserRepository) {
				mockErr := errors.New("error get or insert user")
				userRepo.EXPECT().GetOrInsertByEmail(ctx, &user.User{Email: inputEmail}).Return("", mockErr)
			},
			ExpectErr: errors.New("error get or insert user"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.TODO()
			logger := log.NewNoop()
			mockUserRepo := new(mocks.UserRepository)

			if tc.Setup != nil {
				tc.Setup(ctx, tc.Email, mockUserRepo)
			}

			userSvc := user.NewService(logger, mockUserRepo)

			_, err := userSvc.ValidateUser(ctx, tc.Email)

			assert.Equal(t, tc.ExpectErr, err)
		})
	}
}
