package handlersv1beta1

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/goto/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetLineageGraph(t *testing.T) {
	// TODO[2022-10-13|@sudo-suhas]: Add comprehensive tests
	var (
		userID    = uuid.NewString()
		userEmail = "test@test.com"
		ctx       = user.NewContext(context.Background(), user.User{Email: userEmail})
		logger    = log.NewNoop()
		nodeURN   = "job-1"
		level     = 8
		direction = asset.LineageDirectionUpstream
		ts        = time.Unix(1665659885, 0)
		tspb      = timestamppb.New(ts)
	)
	t.Run("get Lineage", func(t *testing.T) {
		t.Run("should return a graph containing the requested resource, along with it's related resources", func(t *testing.T) {
			lineage := asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "job-1", Target: "table-2"},
					{Source: "table-2", Target: "table-31"},
					{Source: "table-31", Target: "dashboard-30"},
				},
				NodeAttrs: map[string]asset.NodeAttributes{
					"job-1": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "SUCCESS", Timestamp: ts, CreatedAt: ts},
						},
					},
					"table-2": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "FAILED", Timestamp: ts, CreatedAt: ts},
						},
					},
				},
			}
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction, WithAttributes: true}).Return(lineage, nil)
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			got, err := handler.GetGraph(ctx, &compassv1beta1.GetGraphRequest{
				Urn:            nodeURN,
				Level:          uint32(level),
				Direction:      string(direction),
				WithAttributes: proto.Bool(true),
			})
			code := status.Code(err)
			if code != codes.OK {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.OK, code.String())
				return
			}

			expected := &compassv1beta1.GetGraphResponse{
				Data: []*compassv1beta1.LineageEdge{
					{
						Source: "job-1",
						Target: "table-2",
					},
					{
						Source: "table-2",
						Target: "table-31",
					},
					{
						Source: "table-31",
						Target: "dashboard-30",
					},
				},
				NodeAttrs: map[string]*compassv1beta1.GetGraphResponse_NodeAttributes{
					"job-1": {
						Probes: &compassv1beta1.GetGraphResponse_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "SUCCESS", Timestamp: tspb, CreatedAt: tspb},
						},
					},
					"table-2": {
						Probes: &compassv1beta1.GetGraphResponse_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "FAILED", Timestamp: tspb, CreatedAt: tspb},
						},
					},
				},
			}
			if diff := cmp.Diff(got, expected, protocmp.Transform()); diff != "" {
				t.Errorf("expected: %+v\ngot: %+v\ndiff: %s\n", expected, got, diff)
			}
		})

		t.Run("should return error when pass empty user context", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(context.Background(), "").Return("", user.ErrNoUserInformation)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraph(context.Background(), &compassv1beta1.GetGraphRequest{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
			})

			if !strings.Contains(err.Error(), user.ErrNoUserInformation.Error()) {
				t.Errorf("expected error message to be %s, got %s instead", user.ErrNoUserInformation.Error(), err.Error())
			}
		})

		t.Run("should return error when pass invalid direction", func(t *testing.T) {
			invalidDirection := "invalid_direction"

			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraph(ctx, &compassv1beta1.GetGraphRequest{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: invalidDirection,
			})

			code := status.Code(err)
			if code != codes.InvalidArgument {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.InvalidArgument, code.String())
			}

			if !strings.Contains(err.Error(), "invalid direction value") {
				t.Errorf("expected error message to be %s, got %s instead", "invalid direction value", err.Error())
			}
		})

		t.Run("should return error when failed to get lineage", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction, WithAttributes: true}).Return(asset.Lineage{}, fmt.Errorf("failed to get lineage"))
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraph(ctx, &compassv1beta1.GetGraphRequest{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
			})

			code := status.Code(err)
			if code != codes.Internal {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.Internal, code.String())
			}
		})
	})
}

func TestGetLineageGraphV2(t *testing.T) {
	var (
		userID    = uuid.NewString()
		userEmail = "test@test.com"
		ctx       = user.NewContext(context.Background(), user.User{Email: userEmail})
		nodeURN   = "table-1"
		logger    = log.NewNoop()
		level     = 8
		direction = asset.LineageDirectionUpstream
		coverage  = asset.LineageCoverageColumn
		ts        = time.Unix(1665659885, 0)
		tspb      = timestamppb.New(ts)
	)
	t.Run("get Lineage", func(t *testing.T) {
		assetDetail := asset.Asset{
			URN: nodeURN,
		}
		t.Run("should return full asset graph only containing the requested resource, along with it's related resources", func(t *testing.T) {
			lineage := asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "table-1", Target: "table-2"},
					{Source: "table-2", Target: "table-31"},
					{Source: "table-31", Target: "table-30"},
				},
				NodeAttrs: map[string]asset.NodeAttributes{
					"table-1": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "SUCCESS", Timestamp: ts, CreatedAt: ts},
						},
					},
					"table-2": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "FAILED", Timestamp: ts, CreatedAt: ts},
						},
					},
				},
			}
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction, WithAttributes: true}).Return(lineage, nil)
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			got, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
			})
			code := status.Code(err)
			if code != codes.OK {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.OK, code.String())
				return
			}

			expected := &compassv1beta1.GetGraphV2Response{
				Type: string(asset.LineageAssetType),
				Data: []*compassv1beta1.LineageEdgeV2{
					{
						SourceAsset: "table-1",
						TargetAsset: "table-2",
					},
					{
						SourceAsset: "table-2",
						TargetAsset: "table-31",
					},
					{
						SourceAsset: "table-31",
						TargetAsset: "table-30",
					},
				},
				NodeAttrs: map[string]*compassv1beta1.GetGraphV2Response_NodeAttributes{
					"table-1": {
						Probes: &compassv1beta1.GetGraphV2Response_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "SUCCESS", Timestamp: tspb, CreatedAt: tspb},
						},
					},
					"table-2": {
						Probes: &compassv1beta1.GetGraphV2Response_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "FAILED", Timestamp: tspb, CreatedAt: tspb},
						},
					},
				},
			}
			if diff := cmp.Diff(got, expected, protocmp.Transform()); diff != "" {
				t.Errorf("expected: %+v\ngot: %+v\ndiff: %s\n", expected, got, diff)
			}
		})

		t.Run("should return full asset and column level graph containing the requested resource, along with it's related resources", func(t *testing.T) {
			lineage := asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "table-1", SourceColumn: "column-1", Target: "table-2", TargetColumn: "column-2"},
					{Source: "table-2", SourceColumn: "column-2", Target: "table-31", TargetColumn: "column-31"},
					{Source: "table-31", SourceColumn: "column-31", Target: "table-30", TargetColumn: "column-30"},
				},
				NodeAttrs: map[string]asset.NodeAttributes{
					"table-1": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "SUCCESS", Timestamp: ts, CreatedAt: ts},
						},
					},
					"table-2": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "FAILED", Timestamp: ts, CreatedAt: ts},
						},
					},
				},
			}
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetAssetByID(ctx, nodeURN).Return(assetDetail, nil)
			mockSvc.EXPECT().GetColumnLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction, WithAttributes: true, AssetDetail: assetDetail}).Return(lineage, nil)
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			got, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
				Coverage:  proto.String(string(coverage)),
			})
			code := status.Code(err)
			if code != codes.OK {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.OK, code.String())
				return
			}

			expected := &compassv1beta1.GetGraphV2Response{
				Type: string(asset.LineageColumnType),
				Data: []*compassv1beta1.LineageEdgeV2{
					{
						SourceAsset:  "table-1",
						SourceColumn: proto.String("column-1"),
						TargetAsset:  "table-2",
						TargetColumn: proto.String("column-2"),
					},
					{
						SourceAsset:  "table-2",
						SourceColumn: proto.String("column-2"),
						TargetAsset:  "table-31",
						TargetColumn: proto.String("column-31"),
					},
					{
						SourceAsset:  "table-31",
						SourceColumn: proto.String("column-31"),
						TargetAsset:  "table-30",
						TargetColumn: proto.String("column-30"),
					},
				},
				NodeAttrs: map[string]*compassv1beta1.GetGraphV2Response_NodeAttributes{
					"table-1": {
						Probes: &compassv1beta1.GetGraphV2Response_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "SUCCESS", Timestamp: tspb, CreatedAt: tspb},
						},
					},
					"table-2": {
						Probes: &compassv1beta1.GetGraphV2Response_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "FAILED", Timestamp: tspb, CreatedAt: tspb},
						},
					},
				},
			}
			if diff := cmp.Diff(got, expected, protocmp.Transform()); diff != "" {
				t.Errorf("expected: %+v\ngot: %+v\ndiff: %s\n", expected, got, diff)
			}
		})

		t.Run("should return full asset graph for specific column containing the requested resource, along with it's related resources", func(t *testing.T) {
			lineage := asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "table-1", SourceColumn: "column-1", Target: "table-2", TargetColumn: "column-2"},
					{Source: "table-2", SourceColumn: "column-2", Target: "table-31", TargetColumn: "column-31"},
					{Source: "table-31", SourceColumn: "column-31", Target: "table-30", TargetColumn: "column-30"},
				},
				NodeAttrs: map[string]asset.NodeAttributes{
					"table-1": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "SUCCESS", Timestamp: ts, CreatedAt: ts},
						},
					},
					"table-2": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "FAILED", Timestamp: ts, CreatedAt: ts},
						},
					},
				},
			}
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetColumnLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction, WithAttributes: true, TargetColumn: "column-1"}).Return(lineage, nil)
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			got, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:            nodeURN,
				Level:          uint32(level),
				Direction:      string(direction),
				WithAttributes: proto.Bool(true),
				ColumnName:     proto.String("column-1"),
			})
			code := status.Code(err)
			if code != codes.OK {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.OK, code.String())
				return
			}

			expected := &compassv1beta1.GetGraphV2Response{
				Type: string(asset.LineageColumnType),
				Data: []*compassv1beta1.LineageEdgeV2{
					{
						SourceAsset:  "table-1",
						SourceColumn: proto.String("column-1"),
						TargetAsset:  "table-2",
						TargetColumn: proto.String("column-2"),
					},
					{
						SourceAsset:  "table-2",
						SourceColumn: proto.String("column-2"),
						TargetAsset:  "table-31",
						TargetColumn: proto.String("column-31"),
					},
					{
						SourceAsset:  "table-31",
						SourceColumn: proto.String("column-31"),
						TargetAsset:  "table-30",
						TargetColumn: proto.String("column-30"),
					},
				},
				NodeAttrs: map[string]*compassv1beta1.GetGraphV2Response_NodeAttributes{
					"table-1": {
						Probes: &compassv1beta1.GetGraphV2Response_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "SUCCESS", Timestamp: tspb, CreatedAt: tspb},
						},
					},
					"table-2": {
						Probes: &compassv1beta1.GetGraphV2Response_ProbesInfo{
							Latest: &compassv1beta1.Probe{Status: "FAILED", Timestamp: tspb, CreatedAt: tspb},
						},
					},
				},
			}
			if diff := cmp.Diff(got, expected, protocmp.Transform()); diff != "" {
				t.Errorf("expected: %+v\ngot: %+v\ndiff: %s\n", expected, got, diff)
			}
		})

		t.Run("should return error when pass empty user context", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(context.Background(), "").Return("", user.ErrNoUserInformation)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(context.Background(), &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
			})

			if !strings.Contains(err.Error(), user.ErrNoUserInformation.Error()) {
				t.Errorf("expected error message to be %s, got %s instead", user.ErrNoUserInformation.Error(), err.Error())
			}
		})

		t.Run("should return error when pass invalid direction", func(t *testing.T) {
			invalidDirection := "invalid_direction"

			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: invalidDirection,
			})

			code := status.Code(err)
			if code != codes.InvalidArgument {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.InvalidArgument, code.String())
			}

			if !strings.Contains(err.Error(), "invalid direction value") {
				t.Errorf("expected error message to be %s, got %s instead", "invalid direction value", err.Error())
			}
		})

		t.Run("should return error when pass invalid coverage", func(t *testing.T) {
			invalidCoverage := "invalid_coverage"

			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
				Coverage:  proto.String(invalidCoverage),
			})

			code := status.Code(err)
			if code != codes.InvalidArgument {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.InvalidArgument, code.String())
			}

			if !strings.Contains(err.Error(), "invalid coverage value") {
				t.Errorf("expected error message to be %s, got %s instead", "invalid coverage value", err.Error())
			}
		})

		t.Run("should return error when failed to get lineage", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction, WithAttributes: true}).Return(asset.Lineage{}, fmt.Errorf("failed to get lineage"))
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
			})

			code := status.Code(err)
			if code != codes.Internal {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.Internal, code.String())
			}
		})

		t.Run("should return error when invalid asset detail for column lineage", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetAssetByID(ctx, nodeURN).Return(assetDetail, asset.InvalidError{AssetID: nodeURN})
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
				Coverage:  proto.String(string(asset.LineageCoverageColumn)),
			})

			code := status.Code(err)
			if code != codes.InvalidArgument {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.InvalidArgument, code.String())
			}
		})

		t.Run("should return error when get asset detail not found for column lineage", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetAssetByID(ctx, nodeURN).Return(assetDetail, asset.NotFoundError{URN: nodeURN})
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
				Coverage:  proto.String(string(asset.LineageCoverageColumn)),
			})

			code := status.Code(err)
			if code != codes.NotFound {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.NotFound, code.String())
			}
		})

		t.Run("should return error when failed to get asset detail for column lineage", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetAssetByID(ctx, nodeURN).Return(assetDetail, fmt.Errorf("failed to get asset detail"))
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
				Coverage:  proto.String(string(asset.LineageCoverageColumn)),
			})

			code := status.Code(err)
			if code != codes.Internal {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.Internal, code.String())
			}
		})

		t.Run("should return error when failed to get column lineage", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetAssetByID(ctx, nodeURN).Return(assetDetail, nil)
			mockSvc.EXPECT().GetColumnLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction, WithAttributes: true, AssetDetail: assetDetail}).Return(asset.Lineage{}, fmt.Errorf("failed to get column lineage"))
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:       nodeURN,
				Level:     uint32(level),
				Direction: string(direction),
				Coverage:  proto.String(string(asset.LineageCoverageColumn)),
			})

			code := status.Code(err)
			if code != codes.Internal {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.Internal, code.String())
			}
		})

		t.Run("should return error when failed to get column lineage for specific column", func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			mockUserSvc := new(mocks.UserService)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockSvc.EXPECT().GetColumnLineage(ctx, nodeURN, asset.LineageQuery{Level: level, Direction: direction, WithAttributes: true, TargetColumn: "column-1"}).Return(asset.Lineage{}, fmt.Errorf("failed to get column lineage"))
			mockUserSvc.EXPECT().ValidateUser(ctx, userEmail).Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{AssetSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.GetGraphV2(ctx, &compassv1beta1.GetGraphV2Request{
				Urn:        nodeURN,
				Level:      uint32(level),
				Direction:  string(direction),
				ColumnName: proto.String("column-1"),
			})

			code := status.Code(err)
			if code != codes.Internal {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", codes.Internal, code.String())
			}
		})
	})
}
