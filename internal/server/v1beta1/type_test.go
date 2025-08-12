package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/server/v1beta1/mocks"
	"github.com/goto/compass/internal/testutils"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/goto/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetTypes(t *testing.T) {
	var (
		userID    = uuid.NewString()
		userEmail = "test@test.com"
	)
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(tc *testCase, ctx context.Context, as *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GetAllTypesResponse) error
	}

	testCases := []testCase{
		{
			Description:  "should return internal server error if failing to fetch types",
			ExpectStatus: codes.Internal,
			Setup: func(tc *testCase, ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{}, errors.New("failed to fetch type"))
			},
		},
		{
			Description:  "should return internal server error if failing to fetch counts",
			ExpectStatus: codes.Internal,
			Setup: func(tc *testCase, ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{}, errors.New("failed to fetch assets count"))
			},
		},
		{
			Description:  "should return all valid types with its asset count",
			ExpectStatus: codes.OK,
			Setup: func(tc *testCase, ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{
					asset.Type("table"): 10,
					asset.Type("topic"): 30,
					asset.Type("job"):   15,
				}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllTypesResponse) error {
				compare := func(left, right *compassv1beta1.Type) int {
					if left.Name == right.Name && left.Count == right.Count {
						return 0
					}
					if left.Name < right.Name {
						return -1
					}
					return 1
				}

				expected := &compassv1beta1.GetAllTypesResponse{
					Data: []*compassv1beta1.Type{
						{
							Name:  "table",
							Count: 10,
						},
						{
							Name:  "job",
							Count: 15,
						},
						{
							Name:  "dashboard",
							Count: 0,
						},
						{
							Name:  "topic",
							Count: 30,
						},
						{
							Name:  "feature_table",
							Count: 0,
						},
						{
							Name:  "application",
							Count: 0,
						},
						{
							Name:  "model",
							Count: 0,
						},
						{
							Name:  "query",
							Count: 0,
						},
						{
							Name:  "metric",
							Count: 0,
						},
						{
							Name:  "experiment",
							Count: 0,
						},
					},
				}

				if !testutils.AreSlicesEqualIgnoringOrder(resp.Data, expected.Data, compare) {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{Email: userEmail})

			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.AssetService)
			logger := log.NewNoop()
			defer mockSvc.AssertExpectations(t)
			tc.Setup(&tc, ctx, mockSvc)

			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			got, err := handler.GetAllTypes(ctx, &compassv1beta1.GetAllTypesRequest{})
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}
