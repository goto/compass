package asset

import (
	"context"
)

type (
	LineageDirection string
	LineageCoverage  string
	LineageType      string
)

func (dir LineageDirection) IsValid() bool {
	switch dir {
	case LineageDirectionUpstream, LineageDirectionDownstream, "":
		return true
	default:
		return false
	}
}

func (dir LineageCoverage) IsValid() bool {
	switch dir {
	case LineageCoverageAsset, LineageCoverageColumn, "":
		return true
	default:
		return false
	}
}

const (
	LineageDirectionUpstream   LineageDirection = "upstream"
	LineageDirectionDownstream LineageDirection = "downstream"

	LineageCoverageAsset  LineageCoverage = "asset"
	LineageCoverageColumn LineageCoverage = "column"

	LineageAssetType  LineageType = "ASSET_LINEAGE"
	LineageColumnType LineageType = "COLUMN_LINEAGE"
)

type LineageQuery struct {
	Level          int
	Direction      LineageDirection
	WithAttributes bool
	IncludeDeleted bool
	AssetDetail    Asset
	TargetColumn   string
}

//go:generate mockery --name=LineageRepository -r --case underscore --with-expecter --structname=LineageRepository --filename=lineage_repository.go --output=./mocks
type LineageRepository interface {
	GetGraph(ctx context.Context, urn string, query LineageQuery) (LineageGraph, error)
	GetColumnGraph(ctx context.Context, urn string, query LineageQuery) (LineageGraph, error)
	Upsert(ctx context.Context, urn string, upstreams, downstreams []string) error
	UpsertColumnLineage(ctx context.Context, assetURN string, newEdges LineageGraph) error
	DeleteByURN(ctx context.Context, urn string) error
	SoftDeleteByURN(ctx context.Context, urn string) error
	DeleteByURNs(ctx context.Context, urns []string) error
	SoftDeleteByURNs(ctx context.Context, urns []string) error
}

type LineageGraph []LineageEdge

type Lineage struct {
	Edges     []LineageEdge             `json:"edges"`
	NodeAttrs map[string]NodeAttributes `json:"node_attrs"`
}

type LineageEdge struct {
	// Source represents source's node ID
	Source       string `json:"source"`
	SourceColumn string `json:"source_column,omitempty"`

	// Target represents target's node ID
	Target       string `json:"target"`
	TargetColumn string `json:"target_column,omitempty"`

	// Prop is a map containing extra information about the edge
	Prop map[string]interface{} `json:"prop"`
}

type NodeAttributes struct {
	Probes ProbesInfo `json:"probes"`
}

type ProbesInfo struct {
	Latest Probe `json:"latest"`
}
