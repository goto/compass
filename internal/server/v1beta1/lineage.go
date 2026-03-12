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

	withAttributes := req == nil || req.WithAttributes == nil || *req.WithAttributes

	lineage, graphType, err := server.resolveLineageV2(ctx, req, direction, coverage, withAttributes)
	if err != nil {
		return nil, err
	}

	edges, err := buildLineageEdgesV2(lineage.Edges)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	nodeAttrs, err := buildNodeAttrsV2(lineage.NodeAttrs)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetGraphV2Response{
		Type:      string(graphType),
		Data:      edges,
		NodeAttrs: nodeAttrs,
	}, nil
}

func buildLineageEdgesV2(edges []asset.LineageEdge) ([]*compassv1beta1.LineageEdgeV2, error) {
	result := make([]*compassv1beta1.LineageEdgeV2, 0, len(edges))
	for _, edge := range edges {
		edgePB, err := lineageEdgeToProtoV2(edge)
		if err != nil {
			return nil, err
		}
		result = append(result, edgePB)
	}
	return result, nil
}

func buildNodeAttrsV2(nodeAttrs map[string]asset.NodeAttributes) (map[string]*compassv1beta1.GetGraphV2Response_NodeAttributes, error) {
	result := make(map[string]*compassv1beta1.GetGraphV2Response_NodeAttributes, len(nodeAttrs))
	for urn, attrs := range nodeAttrs {
		probesInfo, err := probesInfoToProtoV2(attrs.Probes)
		if err != nil {
			return nil, err
		}
		result[urn] = &compassv1beta1.GetGraphV2Response_NodeAttributes{
			Probes: probesInfo,
		}
	}
	return result, nil
}

func (server *APIServer) resolveLineageV2(
	ctx context.Context,
	req *compassv1beta1.GetGraphV2Request,
	direction asset.LineageDirection,
	coverage asset.LineageCoverage,
	withAttributes bool,
) (asset.Lineage, asset.LineageType, error) {
	baseQuery := asset.LineageQuery{
		Level:          int(req.GetLevel()),
		Direction:      direction,
		WithAttributes: withAttributes,
		IncludeDeleted: req.GetIncludeDeleted(),
	}

	if req != nil && req.ColumnName != nil {
		baseQuery.TargetColumn = req.GetColumnName()
		lineage, err := server.assetService.GetColumnLineage(ctx, req.GetUrn(), baseQuery)
		if err != nil {
			return asset.Lineage{}, "", internalServerError(server.logger, err.Error())
		}
		return lineage, asset.LineageColumnType, nil
	}

	if coverage == asset.LineageCoverageColumn {
		return server.resolveColumnLineageV2(ctx, req.GetUrn(), baseQuery)
	}

	lineage, err := server.assetService.GetLineage(ctx, req.GetUrn(), baseQuery)
	if err != nil {
		return asset.Lineage{}, "", internalServerError(server.logger, err.Error())
	}

	return lineage, asset.LineageAssetType, nil
}

func (server *APIServer) resolveColumnLineageV2(
	ctx context.Context,
	urn string,
	baseQuery asset.LineageQuery,
) (asset.Lineage, asset.LineageType, error) {
	existingAsset, err := server.assetService.GetAssetByID(ctx, urn)
	if err != nil {
		return asset.Lineage{}, "", internalServerError(server.logger, err.Error())
	}

	baseQuery.AssetDetail = existingAsset
	lineage, err := server.assetService.GetColumnLineage(ctx, urn, baseQuery)
	if err != nil {
		return asset.Lineage{}, "", internalServerError(server.logger, err.Error())
	}

	return lineage, asset.LineageColumnType, nil
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
