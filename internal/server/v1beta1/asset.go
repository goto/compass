package handlersv1beta1

//go:generate mockery --name=AssetService -r --case underscore --with-expecter --structname AssetService --filename asset_service.go --output=./mocks
import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/star"
	"github.com/goto/compass/core/user"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/r3labs/diff/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AssetService interface {
	GetAllAssets(ctx context.Context, flt asset.Filter, withTotal bool) ([]asset.Asset, uint32, error)
	GetAssetByID(ctx context.Context, id string) (asset.Asset, error)
	GetAssetByIDWithoutProbes(ctx context.Context, id string) (asset.Asset, error)
	GetAssetByVersion(ctx context.Context, id, version string) (asset.Asset, error)
	GetAssetVersionHistory(ctx context.Context, flt asset.Filter, id string) ([]asset.Asset, error)
	UpsertAsset(ctx context.Context, ast *asset.Asset, upstreams, downstreams []string) (string, error)
	UpsertAssetWithoutLineage(ctx context.Context, ast *asset.Asset) (string, error)
	UpsertPatchAsset(ctx context.Context, ast *asset.Asset, upstreams, downstreams []string, patchData map[string]interface{}) (string, error)
	UpsertPatchAssetWithoutLineage(ctx context.Context, ast *asset.Asset, patchData map[string]interface{}) (string, error)
	DeleteAsset(ctx context.Context, id string) error
	SoftDeleteAsset(ctx context.Context, id, updatedBy string) error
	DeleteAssets(ctx context.Context, request asset.DeleteAssetsRequest) (uint32, error)
	DeleteAssetsByServicesAndUpdatedAt(ctx context.Context, dryRun bool, services string, expiryDuration time.Duration) (uint32, error)
	SoftDeleteAssets(ctx context.Context, request asset.DeleteAssetsRequest, updatedBy string) (uint32, error)

	GetLineage(ctx context.Context, urn string, query asset.LineageQuery) (asset.Lineage, error)
	GetTypes(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error)

	SearchAssets(ctx context.Context, cfg asset.SearchConfig) (results []asset.SearchResult, err error)
	GroupAssets(ctx context.Context, cfg asset.GroupConfig) (results []asset.GroupResult, err error)
	SuggestAssets(ctx context.Context, cfg asset.SearchConfig) (suggestions []string, err error)

	AddProbe(ctx context.Context, assetURN string, probe *asset.Probe) error

	SyncAssets(ctx context.Context, services []string) error
}

func (server *APIServer) GetAllAssets(ctx context.Context, req *compassv1beta1.GetAllAssetsRequest) (*compassv1beta1.GetAllAssetsResponse, error) {
	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	flt, err := asset.NewFilterBuilder().
		Types(req.GetTypes()).
		Services(req.GetServices()).
		Q(req.GetQ()).
		QFields(req.GetQFields()).
		Size(int(req.GetSize())).
		Offset(int(req.GetOffset())).
		SortBy(req.GetSort()).
		SortDirection(req.GetDirection()).
		Data(req.GetData()).
		IsDeleted(req.GetIsDeleted()).
		Build()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	assets, totalCount, err := server.assetService.GetAllAssets(ctx, flt, req.GetWithTotal())
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	assetsProto := make([]*compassv1beta1.Asset, len(assets))
	for i, a := range assets {
		ap, err := assetToProto(a, false)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		assetsProto[i] = ap
	}

	response := &compassv1beta1.GetAllAssetsResponse{
		Data: assetsProto,
	}

	if req.GetWithTotal() {
		response.Total = totalCount
	}

	return response, nil
}

func (server *APIServer) GetAssetByID(ctx context.Context, req *compassv1beta1.GetAssetByIDRequest) (*compassv1beta1.GetAssetByIDResponse, error) {
	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	ast, err := server.assetService.GetAssetByID(ctx, req.GetId())
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	astProto, err := assetToProto(ast, false)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetAssetByIDResponse{
		Data: astProto,
	}, nil
}

func (server *APIServer) GetAssetStargazers(ctx context.Context, req *compassv1beta1.GetAssetStargazersRequest) (*compassv1beta1.GetAssetStargazersResponse, error) {
	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	users, err := server.starService.GetStargazers(ctx, star.Filter{
		Size:   int(req.GetSize()),
		Offset: int(req.GetOffset()),
	}, req.GetId())
	if err != nil {
		if errors.Is(err, star.ErrEmptyUserID) || errors.Is(err, star.ErrEmptyAssetID) || errors.As(err, new(star.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(star.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	usersPB := []*compassv1beta1.User{}
	for _, us := range users {
		usersPB = append(usersPB, userToProto(us))
	}

	return &compassv1beta1.GetAssetStargazersResponse{
		Data: usersPB,
	}, nil
}

func (server *APIServer) GetAssetVersionHistory(ctx context.Context, req *compassv1beta1.GetAssetVersionHistoryRequest) (*compassv1beta1.GetAssetVersionHistoryResponse, error) {
	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	assetVersions, err := server.assetService.GetAssetVersionHistory(ctx, asset.Filter{
		Size:   int(req.GetSize()),
		Offset: int(req.GetOffset()),
	}, req.GetId())
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	assetsPB := []*compassv1beta1.Asset{}
	for _, av := range assetVersions {
		avPB, err := assetToProto(av, true)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		assetsPB = append(assetsPB, avPB)
	}

	return &compassv1beta1.GetAssetVersionHistoryResponse{
		Data: assetsPB,
	}, nil
}

func (server *APIServer) GetAssetByVersion(ctx context.Context, req *compassv1beta1.GetAssetByVersionRequest) (*compassv1beta1.GetAssetByVersionResponse, error) {
	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := asset.ParseVersion(req.GetVersion()); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	ast, err := server.assetService.GetAssetByVersion(ctx, req.GetId(), req.GetVersion())
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	assetPB, err := assetToProto(ast, true)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetAssetByVersionResponse{
		Data: assetPB,
	}, nil
}

func (server *APIServer) UpsertAsset(ctx context.Context, req *compassv1beta1.UpsertAssetRequest) (*compassv1beta1.UpsertAssetResponse, error) {
	userID, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	baseAsset := req.GetAsset()
	if baseAsset == nil {
		return nil, status.Error(codes.InvalidArgument, "asset cannot be empty")
	}

	if err := server.validateUpsertAsset(baseAsset); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ast := server.buildAsset(baseAsset)
	ast.UpdatedBy.ID = userID

	assetID, err := server.upsertAsset(
		ctx,
		ast,
		nil,
		"asset_upsert",
		req.GetUpstreams(),
		req.GetDownstreams(),
	)
	if err != nil {
		return nil, err
	}

	return &compassv1beta1.UpsertAssetResponse{
		Id: assetID,
	}, nil
}

func (server *APIServer) UpsertPatchAsset(ctx context.Context, req *compassv1beta1.UpsertPatchAssetRequest) (*compassv1beta1.UpsertPatchAssetResponse, error) {
	userID, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	baseAsset := req.GetAsset()
	if baseAsset == nil {
		return nil, status.Error(codes.InvalidArgument, "asset cannot be empty")
	}

	err = server.validateUpsertPatchAsset(baseAsset)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	newAsset := asset.Asset{}
	patchAssetMap := decodePatchAssetToMap(baseAsset)
	patchAssetMap["updated_by"] = userID
	newAsset.Patch(patchAssetMap)

	var assetID string
	if len(req.Upstreams) != 0 || len(req.Downstreams) != 0 || req.OverwriteLineage {
		assetID, err = server.upsertAsset(
			ctx, newAsset, patchAssetMap, "asset_upsert_patch", req.GetUpstreams(), req.GetDownstreams(),
		)
	} else {
		assetID, err = server.upsertPatchAssetWithoutLineage(ctx, newAsset, patchAssetMap)
	}
	if err != nil {
		return nil, err
	}

	return &compassv1beta1.UpsertPatchAssetResponse{
		Id: assetID,
	}, nil
}

// DeleteAsset can accept ID or URN of asset
func (server *APIServer) DeleteAsset(ctx context.Context, req *compassv1beta1.DeleteAssetRequest) (*compassv1beta1.DeleteAssetResponse, error) {
	userID, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = server.assetService.SoftDeleteAsset(ctx, req.GetId(), userID)
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.DeleteAssetResponse{}, nil
}

func (server *APIServer) DeleteAssets(ctx context.Context, req *compassv1beta1.DeleteAssetsRequest) (*compassv1beta1.DeleteAssetsResponse, error) {
	var affectedRows uint32
	userID, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		server.logger.Warn("delete assets by query",
			"affected rows", affectedRows,
			"query delete", req.QueryExpr,
			"dry run", req.DryRun)
	}()

	deleteAssetsRequest := asset.DeleteAssetsRequest{
		QueryExpr: req.QueryExpr,
		DryRun:    req.DryRun,
	}
	affectedRows, err = server.assetService.SoftDeleteAssets(ctx, deleteAssetsRequest, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &compassv1beta1.DeleteAssetsResponse{
		AffectedRows: affectedRows,
	}, nil
}

func (server *APIServer) CreateAssetProbe(ctx context.Context, req *compassv1beta1.CreateAssetProbeRequest) (*compassv1beta1.CreateAssetProbeResponse, error) {
	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if req.Probe.Id != "" && !isValidUUID(req.Probe.Id) {
		return nil, status.Error(codes.InvalidArgument, "id should be a valid UUID")
	}
	if req.Probe.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "Status is required")
	}
	if !req.Probe.Timestamp.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "Timestamp is required")
	}

	probe := asset.Probe{
		ID:           req.Probe.Id,
		Status:       req.Probe.Status,
		StatusReason: req.Probe.StatusReason,
		Metadata:     req.Probe.Metadata.AsMap(),
		Timestamp:    req.Probe.Timestamp.AsTime(),
	}
	if err := server.assetService.AddProbe(ctx, req.AssetUrn, &probe); err != nil {
		switch {
		case errors.As(err, &asset.NotFoundError{}):
			return nil, status.Error(codes.NotFound, err.Error())

		case errors.Is(err, asset.ErrProbeExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &compassv1beta1.CreateAssetProbeResponse{
		Id: probe.ID,
	}, nil
}

func (server *APIServer) SyncAssets(ctx context.Context, req *compassv1beta1.SyncAssetsRequest) (*compassv1beta1.SyncAssetsResponse, error) {
	if err := server.assetService.SyncAssets(ctx, req.GetServices()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &compassv1beta1.SyncAssetsResponse{}, nil
}

func (server *APIServer) upsertAsset(
	ctx context.Context,
	ast asset.Asset,
	patchData map[string]interface{},
	mode string,
	reqUpstreams,
	reqDownstreams []*compassv1beta1.LineageNode,
) (id string, err error) {
	defer func() {
		server.assetUpdateCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("compass.update_method", mode),
			attribute.String("asset.type", (string)(ast.Type)),
			attribute.String("asset.service", ast.Service),
			attribute.Bool("operation.success", err == nil),
		))
	}()

	upstreams := make([]string, 0, len(reqUpstreams))
	for _, pb := range reqUpstreams {
		upstreams = append(upstreams, pb.Urn)
	}
	downstreams := make([]string, 0, len(reqDownstreams))
	for _, pb := range reqDownstreams {
		downstreams = append(downstreams, pb.Urn)
	}

	var assetID string
	if patchData == nil {
		assetID, err = server.assetService.UpsertAsset(ctx, &ast, upstreams, downstreams)
	} else {
		assetID, err = server.assetService.UpsertPatchAsset(ctx, &ast, upstreams, downstreams, patchData)
	}

	if err != nil {
		switch {
		case errors.As(err, new(asset.InvalidError)):
			return "", status.Error(codes.InvalidArgument, err.Error())
		}
		return "", internalServerError(server.logger, err.Error())
	}

	return assetID, nil
}

func (server *APIServer) upsertPatchAssetWithoutLineage(ctx context.Context, ast asset.Asset, patchData map[string]interface{}) (id string, err error) {
	const mode = "asset_upsert_patch_without_lineage"
	defer func() {
		server.assetUpdateCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("compass.update_method", mode),
			attribute.String("asset.type", (string)(ast.Type)),
			attribute.String("asset.service", ast.Service),
			attribute.Bool("operation.success", err == nil),
		))
	}()

	assetID, err := server.assetService.UpsertPatchAssetWithoutLineage(ctx, &ast, patchData)
	if err != nil {
		switch {
		case errors.As(err, new(asset.InvalidError)):
			return "", status.Error(codes.InvalidArgument, err.Error())
		}

		return "", internalServerError(server.logger, err.Error())
	}

	return assetID, nil
}

func (server *APIServer) buildAsset(baseAsset *compassv1beta1.UpsertAssetRequest_Asset) asset.Asset {
	ast := asset.Asset{
		URN:         baseAsset.GetUrn(),
		Service:     baseAsset.GetService(),
		Type:        asset.Type(baseAsset.GetType()),
		Name:        baseAsset.GetName(),
		Description: baseAsset.GetDescription(),
		Data:        baseAsset.GetData().AsMap(),
		URL:         baseAsset.Url,
		Labels:      baseAsset.GetLabels(),
	}

	var owners []user.User
	for _, owner := range dedupe(baseAsset.GetOwners()) {
		owners = append(owners, user.User{
			ID:       owner.Id,
			Email:    owner.Email,
			Provider: owner.Provider,
		})
	}
	ast.Owners = owners

	return ast
}

func (*APIServer) validateUpsertAsset(ast *compassv1beta1.UpsertAssetRequest_Asset) error {
	if urn := ast.GetUrn(); urn == "" {
		return fmt.Errorf("urn is required and can't be empty")
	}

	typ := ast.GetType()
	if typ == "" {
		return fmt.Errorf("type is required and can't be empty")
	}

	if !asset.Type(typ).IsValid() {
		return fmt.Errorf("type is invalid")
	}

	if name := ast.GetName(); name == "" {
		return fmt.Errorf("name is required and can't be empty")
	}

	if data := ast.GetData(); data == nil {
		return fmt.Errorf("name is required and can't be nil")
	}

	if service := ast.GetService(); service == "" {
		return fmt.Errorf("service is required and can't be empty")
	}

	return nil
}

func (*APIServer) validateUpsertPatchAsset(ast *compassv1beta1.UpsertPatchAssetRequest_Asset) error {
	if urn := ast.GetUrn(); urn == "" {
		return fmt.Errorf("urn is required and can't be empty")
	}

	typ := ast.GetType()
	if typ == "" {
		return fmt.Errorf("type is required and can't be empty")
	}

	if !asset.Type(typ).IsValid() {
		return fmt.Errorf("type is invalid")
	}

	if service := ast.GetService(); service == "" {
		return fmt.Errorf("service is required and can't be empty")
	}

	return nil
}

func decodePatchAssetToMap(pb *compassv1beta1.UpsertPatchAssetRequest_Asset) map[string]interface{} {
	if pb == nil {
		return nil
	}
	m := map[string]interface{}{}
	m["urn"] = pb.GetUrn()
	m["type"] = pb.GetType()
	m["service"] = pb.GetService()
	if pb.GetName() != nil {
		m["name"] = pb.GetName().Value
	}
	if pb.GetDescription() != nil {
		m["description"] = pb.GetDescription().Value
	}
	if pb.GetData() != nil {
		m["data"] = pb.GetData().AsMap()
	}
	if len(pb.Url) > 0 {
		m["url"] = pb.Url
	}
	if pb.GetLabels() != nil {
		m["labels"] = pb.GetLabels()
	}
	if len(pb.GetOwners()) > 0 {
		ownersMap := []map[string]interface{}{}
		ownersPB := dedupe(pb.GetOwners())
		for _, ownerPB := range ownersPB {
			ownerMap := map[string]interface{}{}
			if ownerPB.GetId() != "" {
				ownerMap["id"] = ownerPB.GetId()
			}
			if ownerPB.GetUuid() != "" {
				ownerMap["uuid"] = ownerPB.GetUuid()
			}
			if ownerPB.GetEmail() != "" {
				ownerMap["email"] = ownerPB.GetEmail()
			}
			if ownerPB.GetProvider() != "" {
				ownerMap["provider"] = ownerPB.GetProvider()
			}
			ownersMap = append(ownersMap, ownerMap)
		}
		m["owners"] = ownersMap
	}

	return m
}

func dedupe(owners []*compassv1beta1.User) []*compassv1beta1.User {
	n := len(owners)
	uniq := make([]*compassv1beta1.User, 0, n)
	ids := make(map[string]struct{}, n)
	emails := make(map[string]struct{}, n)
	for _, o := range owners {
		if _, ok := ids[o.Id]; ok {
			continue
		}
		if _, ok := emails[o.Email]; ok {
			continue
		}
		if o.Id != "" {
			ids[o.Id] = struct{}{}
		}
		if o.Email != "" {
			emails[o.Email] = struct{}{}
		}
		uniq = append(uniq, o)
	}
	return uniq
}

// assetToProto transforms struct to proto
func assetToProto(a asset.Asset, withChangelog bool) (*compassv1beta1.Asset, error) {
	var data *structpb.Struct
	if len(a.Data) > 0 {
		var err error
		data, err = structpb.NewStruct(a.Data)
		if err != nil {
			return nil, err
		}
	}

	var owners []*compassv1beta1.User
	for _, o := range a.Owners {
		owners = append(owners, userToProto(o))
	}

	var changelog []*compassv1beta1.Change
	if withChangelog {
		var err error
		changelog, err = changelogToProto(a.Changelog)
		if err != nil {
			return nil, err
		}
	}

	var createdAt *timestamppb.Timestamp
	if !a.CreatedAt.IsZero() {
		createdAt = timestamppb.New(a.CreatedAt)
	}

	var updatedAt *timestamppb.Timestamp
	if !a.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(a.UpdatedAt)
	}

	var probes []*compassv1beta1.Probe
	for _, probe := range a.Probes {
		probeProto, err := probeToProto(probe)
		if err != nil {
			return nil, fmt.Errorf("convert probe to proto: %w", err)
		}

		probes = append(probes, probeProto)
	}

	return &compassv1beta1.Asset{
		Id:          a.ID,
		Urn:         a.URN,
		Type:        string(a.Type),
		Service:     a.Service,
		Name:        a.Name,
		Description: a.Description,
		Data:        data,
		Url:         a.URL,
		Labels:      a.Labels,
		Owners:      owners,
		Version:     a.Version,
		UpdatedBy:   userToProto(a.UpdatedBy),
		Changelog:   changelog,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Probes:      probes,
		IsDeleted:   a.IsDeleted,
	}, nil
}

// probeToProto transforms asset.Probe struct to proto
func probeToProto(probe asset.Probe) (*compassv1beta1.Probe, error) {
	res := &compassv1beta1.Probe{
		Id:           probe.ID,
		AssetUrn:     probe.AssetURN,
		Status:       probe.Status,
		StatusReason: probe.StatusReason,
		Timestamp:    timestamppb.New(probe.Timestamp),
		CreatedAt:    timestamppb.New(probe.CreatedAt),
	}

	if probe.Metadata != nil {
		m, err := structpb.NewStruct(probe.Metadata)
		if err != nil {
			return nil, fmt.Errorf("error creating probe metadata: %w", err)
		}

		res.Metadata = m
	}

	return res, nil
}

// changelogToProto transforms changelog struct to proto
func changelogToProto(cl diff.Changelog) ([]*compassv1beta1.Change, error) {
	if len(cl) == 0 {
		return nil, nil
	}
	var protoChanges []*compassv1beta1.Change
	for _, ch := range cl {
		chProto, err := diffChangeToProto(ch)
		if err != nil {
			return nil, err
		}

		protoChanges = append(protoChanges, chProto)
	}
	return protoChanges, nil
}

func diffChangeToProto(dc diff.Change) (*compassv1beta1.Change, error) {
	from, err := structpb.NewValue(dc.From)
	if err != nil {
		return nil, err
	}
	to, err := structpb.NewValue(dc.To)
	if err != nil {
		return nil, err
	}

	return &compassv1beta1.Change{
		Type: dc.Type,
		Path: dc.Path,
		From: from,
		To:   to,
	}, nil
}

// assetFromProto transforms proto to struct
// changelog is not populated by user, it should always be processed and coming from the server
func assetFromProto(pb *compassv1beta1.Asset) asset.Asset {
	var assetOwners []user.User
	for _, op := range pb.GetOwners() {
		if op == nil {
			continue
		}
		assetOwners = append(assetOwners, userFromProto(op))
	}

	var dataValue map[string]interface{}
	if pb.GetData() != nil {
		dataValue = pb.GetData().AsMap()
	}

	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var updatedBy user.User
	if pb.GetUpdatedBy() != nil {
		updatedBy = userFromProto(pb.GetUpdatedBy())
	}

	var clog diff.Changelog
	if len(pb.GetChangelog()) > 0 {
		for _, cg := range pb.GetChangelog() {
			if cg == nil {
				continue
			}
			clog = append(clog, diffChangeFromProto(cg))
		}
	}

	return asset.Asset{
		ID:          pb.GetId(),
		URN:         pb.GetUrn(),
		Type:        asset.Type(pb.GetType()),
		Service:     pb.GetService(),
		Name:        pb.GetName(),
		Description: pb.GetDescription(),
		Data:        dataValue,
		Labels:      pb.GetLabels(),
		Owners:      assetOwners,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Version:     pb.GetVersion(),
		Changelog:   clog,
		UpdatedBy:   updatedBy,
		IsDeleted:   pb.GetIsDeleted(),
	}
}

// diffChangeFromProto converts Change proto to diff.Change
func diffChangeFromProto(pb *compassv1beta1.Change) diff.Change {
	var fromItf interface{}
	if pb.GetFrom() != nil {
		fromItf = pb.GetFrom().AsInterface()
	}

	var toItf interface{}
	if pb.GetTo() != nil {
		toItf = pb.GetTo().AsInterface()
	}

	return diff.Change{
		Type: pb.GetType(),
		Path: pb.GetPath(),
		From: fromItf,
		To:   toItf,
	}
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
