package handlersv1beta1

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/server/v1beta1/mocks"
	"github.com/goto/compass/internal/testutils"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/goto/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestSearch(t *testing.T) {
	var (
		userID    = uuid.NewString()
		userEmail = "test@test.com"
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.SearchAssetsRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.SearchAssetsResponse) error
	}

	testCases := []testCase{
		{
			Description: "should report internal server if asset searcher fails",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text: "test",
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				err := fmt.Errorf("service unavailable")
				as.EXPECT().SearchAssets(ctx, mock.AnythingOfType("asset.SearchConfig")).
					Return([]asset.SearchResult{}, err)
			},
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should pass filter to search config format",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text: "resource",
				Filter: map[string]string{
					"data.landscape": "th",
					"type":           "topic",
					"service":        "kafka,rabbitmq",
				},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.SearchConfig{
					Text: "resource",
					Filters: map[string][]string{
						"type":           {"topic"},
						"service":        {"kafka", "rabbitmq"},
						"data.landscape": {"th"},
					},
				}

				as.EXPECT().SearchAssets(ctx, cfg).Return([]asset.SearchResult{}, nil)
			},
		},
		{
			Description: "should pass queries to search config format",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text: "resource",
				Filter: map[string]string{
					"data.landscape": "th",
					"type":           "topic",
					"service":        "kafka,rabbitmq",
				},
				Query: map[string]string{
					"data.columns.name": "timestamp",
					"owners.email":      "john.doe@email.com",
				},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.SearchConfig{
					Text: "resource",
					Filters: map[string][]string{
						"type":           {"topic"},
						"service":        {"kafka", "rabbitmq"},
						"data.landscape": {"th"},
					},
					Queries: map[string]string{
						"data.columns.name": "timestamp",
						"owners.email":      "john.doe@email.com",
					},
				}

				as.EXPECT().SearchAssets(ctx, cfg).Return([]asset.SearchResult{}, nil)
			},
		},
		{
			Description: "should parse offset",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text:   "resource",
				Offset: 10,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.SearchConfig{
					Text:    "resource",
					Filters: make(map[string][]string),
					Queries: map[string]string(nil),
					Offset:  10,
				}

				as.EXPECT().SearchAssets(ctx, cfg).Return([]asset.SearchResult{}, nil)
			},
		},
		{
			Description: "should return the matched documents",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text: "test",
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.SearchConfig{
					Text:    "test",
					Filters: make(map[string][]string),
					Queries: map[string]string(nil),
				}
				response := []asset.SearchResult{
					{
						Type:        "test",
						ID:          "test-resource",
						Description: "some description",
						Service:     "test-service",
						Labels: map[string]string{
							"entity":    "gotocompany",
							"landscape": "id",
						},
					},
				}
				as.EXPECT().SearchAssets(ctx, cfg).Return(response, nil)
			},
			PostCheck: func(resp *compassv1beta1.SearchAssetsResponse) error {
				expected := &compassv1beta1.SearchAssetsResponse{
					Data: []*compassv1beta1.Asset{
						{
							Id:          "test-resource",
							Description: "some description",
							Service:     "test-service",
							Type:        "test",
							Labels: map[string]string{
								"entity":    "gotocompany",
								"landscape": "id",
							},
						},
					},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
		{
			Description: "should return the requested number of assets",
			Request: &compassv1beta1.SearchAssetsRequest{
				Text: "resource",
				Size: 10,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.SearchConfig{
					Text:       "resource",
					MaxResults: 10,
					Filters:    make(map[string][]string),
					Queries:    map[string]string(nil),
				}

				var results []asset.SearchResult
				for i := 0; i < cfg.MaxResults; i++ {
					urn := fmt.Sprintf("resource-%d", i+1)
					name := fmt.Sprintf("resource %d", i+1)
					r := asset.SearchResult{
						ID:          urn,
						Type:        "table",
						Description: name,
						Service:     "kafka",
						Labels: map[string]string{
							"landscape": "id",
							"entity":    "gotocompany",
						},
					}

					results = append(results, r)
				}

				as.EXPECT().SearchAssets(ctx, cfg).Return(results, nil)
			},
			PostCheck: func(resp *compassv1beta1.SearchAssetsResponse) error {
				expectedSize := 10
				actualSize := len(resp.Data)
				if expectedSize != actualSize {
					return fmt.Errorf("expected search request to return %d results, returned %d results instead", expectedSize, actualSize)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{Email: userEmail})

			logger := log.NewNoop()
			mockUserSvc := mocks.NewUserService(t)
			mockSvc := mocks.NewAssetService(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}

			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			got, err := handler.SearchAssets(ctx, tc.Request)
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

func TestSuggest(t *testing.T) {
	var (
		userID    = uuid.NewString()
		userEmail = "test@test.com"
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.SuggestAssetsRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.SuggestAssetsResponse) error
	}

	testCases := []testCase{
		{
			Description:  "should return invalid arguments if 'text' parameter is empty or missing",
			ExpectStatus: codes.InvalidArgument,
			Request:      &compassv1beta1.SuggestAssetsRequest{},
		},
		{
			Description: "should report internal server error if searcher fails",
			Request: &compassv1beta1.SuggestAssetsRequest{
				Text: "test",
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.SearchConfig{
					Text: "test",
				}
				as.EXPECT().SuggestAssets(ctx, cfg).Return([]string{}, fmt.Errorf("service unavailable"))
			},
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return suggestions",
			Request: &compassv1beta1.SuggestAssetsRequest{
				Text: "test",
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.SearchConfig{
					Text: "test",
				}
				response := []string{
					"test",
					"test2",
					"t est",
					"t_est",
				}

				as.EXPECT().SuggestAssets(ctx, cfg).Return(response, nil)
			},
			PostCheck: func(resp *compassv1beta1.SuggestAssetsResponse) error {
				expected := &compassv1beta1.SuggestAssetsResponse{
					Data: []string{
						"test",
						"test2",
						"t est",
						"t_est",
					},
				}
				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{Email: userEmail})

			logger := log.NewNoop()
			mockUserSvc := mocks.NewUserService(t)
			mockSvc := mocks.NewAssetService(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}

			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			got, err := handler.SuggestAssets(ctx, tc.Request)
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

func TestGroupAssets(t *testing.T) {
	var (
		userID    = uuid.NewString()
		userEmail = "test@test.com"
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GroupAssetsRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GroupAssetsResponse)
	}

	testCases := []testCase{
		{
			Description:  "should return invalid argument if 'groupby' parameter is empty or missing",
			ExpectStatus: codes.InvalidArgument,
			Request:      &compassv1beta1.GroupAssetsRequest{},
		},
		{
			Description: "should report internal server if asset grouper fails",
			Request: &compassv1beta1.GroupAssetsRequest{
				Groupby: []string{"groupby"},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				err := fmt.Errorf("service unavailable")
				as.EXPECT().GroupAssets(ctx, mock.AnythingOfType("asset.GroupConfig")).
					Return([]asset.GroupResult{}, err)
			},
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should pass filter to group config format",
			Request: &compassv1beta1.GroupAssetsRequest{
				Groupby: []string{"resource"},
				Filter: map[string]string{
					"data.landscape": "th",
					"type":           "topic",
					"service":        "kafka,rabbitmq",
				},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.GroupConfig{
					GroupBy: []string{"resource"},
					Filters: map[string][]string{
						"type":           {"topic"},
						"service":        {"kafka", "rabbitmq"},
						"data.landscape": {"th"},
					},
				}
				as.EXPECT().GroupAssets(ctx, cfg).Return([]asset.GroupResult{}, nil)
			},
		},
		{
			Description: "should pass include fields to search config format",
			Request: &compassv1beta1.GroupAssetsRequest{
				Groupby: []string{"resource"},
				Filter: map[string]string{
					"data.landscape": "th",
					"type":           "topic",
					"service":        "kafka,rabbitmq",
				},
				IncludeFields: []string{"data.columns.name", "owners.email"},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.GroupConfig{
					GroupBy: []string{"resource"},
					Filters: map[string][]string{
						"type":           {"topic"},
						"service":        {"kafka", "rabbitmq"},
						"data.landscape": {"th"},
					},
					IncludedFields: []string{"data.columns.name", "owners.email"},
				}
				as.EXPECT().GroupAssets(ctx, cfg).Return([]asset.GroupResult{}, nil)
			},
		},
		{
			Description: "should return the grouped documents",
			Request: &compassv1beta1.GroupAssetsRequest{
				Groupby: []string{"resource"},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.GroupConfig{
					GroupBy: []string{"resource"},
					Filters: make(map[string][]string),
				}
				response := []asset.GroupResult{
					{
						Fields: []asset.GroupField{
							{
								Name:  "resource",
								Value: "kafka",
							},
						},
						Assets: []asset.Asset{
							{
								Type:        "test",
								ID:          "test-resource",
								Description: "some description",
								Service:     "test-service",
								Labels: map[string]string{
									"entity":    "gotocompany",
									"landscape": "id",
								},
							},
						},
					},
				}
				as.EXPECT().GroupAssets(ctx, cfg).Return(response, nil)
			},

			PostCheck: func(resp *compassv1beta1.GroupAssetsResponse) {
				expected := &compassv1beta1.GroupAssetsResponse{
					AssetGroups: []*compassv1beta1.AssetGroup{
						{
							GroupFields: []*compassv1beta1.GroupField{
								{
									GroupKey:   "resource",
									GroupValue: "kafka",
								},
							},
							Assets: []*compassv1beta1.Asset{
								{
									Id:          "test-resource",
									Description: "some description",
									Service:     "test-service",
									Type:        "test",
									Labels: map[string]string{
										"entity":    "gotocompany",
										"landscape": "id",
									},
								},
							},
						},
					},
				}
				testutils.AssertEqualProto(t, expected, resp)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{Email: userEmail})

			logger := log.NewNoop()
			mockUserSvc := mocks.NewUserService(t)
			mockSvc := mocks.NewAssetService(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}

			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			expected, err := handler.GroupAssets(ctx, tc.Request)
			code := status.Code(err)
			assert.Equal(t, tc.ExpectStatus, code)
			if tc.PostCheck != nil {
				tc.PostCheck(expected)
			}
		})
	}
}
