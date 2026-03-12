package handlersv1beta1

import (
	"context"
	"fmt"

	"github.com/goto/compass/core/asset"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (server *APIServer) GetGraph(ctx context.Context, req *compassv1beta1.GetGraphRequest) (*compassv1beta1.GetGraphResponse, error) {
	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	direction := asset.LineageDirection(req.GetDirection())
	if !direction.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid direction value")
	}

	withAttributes := true
	if req != nil && req.WithAttributes != nil {
		withAttributes = *req.WithAttributes
	}

	lineage, err := server.assetService.GetLineage(ctx, req.GetUrn(), asset.LineageQuery{
		Level:          int(req.GetLevel()),
		Direction:      direction,
		WithAttributes: withAttributes,
		IncludeDeleted: req.GetIncludeDeleted(),
	})
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	edges := make([]*compassv1beta1.LineageEdge, 0, len(lineage.Edges))
	for _, edge := range lineage.Edges {
		edgePB, err := lineageEdgeToProto(edge)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		edges = append(edges, edgePB)
	}

	nodeAttrs := make(map[string]*compassv1beta1.GetGraphResponse_NodeAttributes, len(lineage.NodeAttrs))
	for urn, attrs := range lineage.NodeAttrs {
		probesInfo, err := probesInfoToProto(attrs.Probes)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}

		nodeAttrs[urn] = &compassv1beta1.GetGraphResponse_NodeAttributes{
			Probes: probesInfo,
		}
	}

	return &compassv1beta1.GetGraphResponse{
		Data:      edges,
		NodeAttrs: nodeAttrs,
	}, nil
}

func lineageEdgeToProto(e asset.LineageEdge) (*compassv1beta1.LineageEdge, error) {
	var (
		propPB *structpb.Struct
		err    error
	)

	if len(e.Prop) > 0 {
		propPB, err = structpb.NewStruct(e.Prop)
		if err != nil {
			return nil, err
		}
	}
	return &compassv1beta1.LineageEdge{
		Source: e.Source,
		Target: e.Target,
		Prop:   propPB,
	}, nil
}

func probesInfoToProto(probes asset.ProbesInfo) (*compassv1beta1.GetGraphResponse_ProbesInfo, error) {
	latest, err := probeToProto(probes.Latest)
	if err != nil {
		return nil, fmt.Errorf("convert probe to proto representation: %w", err)
	}

	return &compassv1beta1.GetGraphResponse_ProbesInfo{
		Latest: latest,
	}, nil
}

func (server *APIServer) GetGraphV2(ctx context.Context, req *compassv1beta1.GetGraphV2Request) (*compassv1beta1.GetGraphV2Response, error) {
	var (
		lineage   asset.Lineage
		graphType asset.LineageType
	)

	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	direction := asset.LineageDirection(req.GetDirection())
	if !direction.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid direction value")
	}

	coverage := asset.LineageCoverage(req.GetCoverage())
	if !coverage.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid coverage value")
	}

	withAttributes := true
	if req != nil && req.WithAttributes != nil {
		withAttributes = *req.WithAttributes
	}

	if req != nil && req.ColumnName != nil {
		graphType = asset.LineageColumnType
		lineage, err = server.assetService.GetColumnLineage(ctx, req.GetUrn(), asset.LineageQuery{
			Level:          int(req.GetLevel()),
			Direction:      direction,
			WithAttributes: withAttributes,
			IncludeDeleted: req.GetIncludeDeleted(),
			TargetColumn:   req.GetColumnName(),
		})
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
	} else {
		switch coverage {
		case asset.LineageCoverageColumn:
			graphType = asset.LineageColumnType
			existingAsset, err := server.assetService.GetAssetByID(ctx, req.GetUrn())
			if err != nil {
				return nil, internalServerError(server.logger, err.Error())
			}

			lineage, err = server.assetService.GetColumnLineage(ctx, req.GetUrn(), asset.LineageQuery{
				Level:          int(req.GetLevel()),
				Direction:      direction,
				WithAttributes: withAttributes,
				IncludeDeleted: req.GetIncludeDeleted(),
				AssetDetail:    existingAsset,
			})
			if err != nil {
				return nil, internalServerError(server.logger, err.Error())
			}
		default:
			graphType = asset.LineageAssetType
			lineage, err = server.assetService.GetLineage(ctx, req.GetUrn(), asset.LineageQuery{
				Level:          int(req.GetLevel()),
				Direction:      direction,
				WithAttributes: withAttributes,
				IncludeDeleted: req.GetIncludeDeleted(),
			})
			if err != nil {
				return nil, internalServerError(server.logger, err.Error())
			}
		}
	}

	edges := make([]*compassv1beta1.LineageEdgeV2, 0, len(lineage.Edges))
	for _, edge := range lineage.Edges {
		edgePB, err := lineageEdgeToProtoV2(edge)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		edges = append(edges, edgePB)
	}

	nodeAttrs := make(map[string]*compassv1beta1.GetGraphV2Response_NodeAttributes, len(lineage.NodeAttrs))
	for urn, attrs := range lineage.NodeAttrs {
		probesInfo, err := probesInfoToProtoV2(attrs.Probes)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}

		nodeAttrs[urn] = &compassv1beta1.GetGraphV2Response_NodeAttributes{
			Probes: probesInfo,
		}
	}

	return &compassv1beta1.GetGraphV2Response{
		Type:      string(graphType),
		Data:      edges,
		NodeAttrs: nodeAttrs,
	}, nil
}

func lineageEdgeToProtoV2(e asset.LineageEdge) (*compassv1beta1.LineageEdgeV2, error) {
	var (
		propPB *structpb.Struct
		err    error
	)

	if len(e.Prop) > 0 {
		propPB, err = structpb.NewStruct(e.Prop)
		if err != nil {
			return nil, err
		}
	}
	edge := &compassv1beta1.LineageEdgeV2{
		SourceAsset: e.Source,
		TargetAsset: e.Target,
		Prop:        propPB,
	}
	if e.SourceColumn != "" {
		edge.SourceColumn = &e.SourceColumn
	}
	if e.TargetColumn != "" {
		edge.TargetColumn = &e.TargetColumn
	}
	return edge, nil
}

func probesInfoToProtoV2(probes asset.ProbesInfo) (*compassv1beta1.GetGraphV2Response_ProbesInfo, error) {
	latest, err := probeToProto(probes.Latest)
	if err != nil {
		return nil, fmt.Errorf("convert probe to proto representation: %w", err)
	}

	return &compassv1beta1.GetGraphV2Response_ProbesInfo{
		Latest: latest,
	}, nil
}
