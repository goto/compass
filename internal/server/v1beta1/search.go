package handlersv1beta1

import (
	"context"
	"fmt"
	"strings"

	"github.com/goto/compass/core/asset"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *APIServer) SearchAssets(ctx context.Context, req *compassv1beta1.SearchAssetsRequest) (*compassv1beta1.SearchAssetsResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(req.GetText())
	if text == "" {
		return nil, status.Error(codes.InvalidArgument, "'text' must be specified")
	}

	cfg := asset.SearchConfig{
		Text:       text,
		MaxResults: int(req.GetSize()),
		Filters:    filterConfigFromValues(req.GetFilter()),
		RankBy:     req.GetRankby(),
		Queries:    req.GetQuery(),
	}

	results, err := server.assetService.SearchAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error searching asset: %s", err.Error()))
	}

	assetsPB := []*compassv1beta1.Asset{}
	for _, sr := range results {

		assetPB, err := assetToProto(sr.ToAsset(), false)
		if err != nil {
			return nil, internalServerError(server.logger, fmt.Sprintf("error converting assets to proto: %s", err.Error()))
		}
		assetsPB = append(assetsPB, assetPB)
	}

	return &compassv1beta1.SearchAssetsResponse{
		Data: assetsPB,
	}, nil
}

func (server *APIServer) GroupAssets(ctx context.Context, req *compassv1beta1.GroupAssetsRequest) (*compassv1beta1.GroupAssetsResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	groupby := req.GetGroupby()
	if len(groupby) == 0 || groupby[0] == "" {
		return nil, status.Error(codes.InvalidArgument, "'groupby' must be specified")
	}

	cfg := asset.GroupConfig{
		GroupBy:        groupby,
		Filters:        filterConfigFromValues(req.GetFilter()),
		IncludedFields: req.GetIncludeFields(),
		Size:           int(req.GetSize()),
		Logger:         server.logger,
	}

	results, err := server.assetService.GroupAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error searching asset: %s", err.Error()))
	}

	groupInfoArr := make([]*compassv1beta1.AssetGroup, len(results))
	for idx, gr := range results {
		assetsPB := make([]*compassv1beta1.Asset, len(gr.Assets))
		for assetIdx, as := range gr.Assets {
			assetPB, err := assetToProto(as, false)
			if err != nil {
				return nil, internalServerError(server.logger, fmt.Sprintf("error converting assets to proto: %s", err.Error()))
			}
			assetsPB[assetIdx] = assetPB
		}

		groupInfo := &compassv1beta1.AssetGroup{
			GroupFields: []*compassv1beta1.GroupField{
				{
					GroupKey:   cfg.GroupBy[0],
					GroupValue: gr.Key,
				},
			},
			Assets: assetsPB,
		}
		groupInfoArr[idx] = groupInfo
	}

	return &compassv1beta1.GroupAssetsResponse{
		AssetGroups: groupInfoArr,
	}, nil
}

func (server *APIServer) SuggestAssets(ctx context.Context, req *compassv1beta1.SuggestAssetsRequest) (*compassv1beta1.SuggestAssetsResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(req.GetText())
	if text == "" {
		return nil, status.Error(codes.InvalidArgument, "'text' must be specified")
	}

	cfg := asset.SearchConfig{
		Text: text,
	}

	suggestions, err := server.assetService.SuggestAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.SuggestAssetsResponse{
		Data: suggestions,
	}, nil
}

func filterConfigFromValues(fltMap map[string]string) map[string][]string {
	var filter = make(map[string][]string)
	for key, value := range fltMap {
		var filterValues []string
		filterValues = append(filterValues, strings.Split(value, ",")...)

		filter[key] = filterValues
	}
	return filter
}
