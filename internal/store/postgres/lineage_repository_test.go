package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/salt/log"
	"github.com/stretchr/testify/suite"
)

type LineageRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	repository *postgres.LineageRepository
}

func (r *LineageRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus()
	client, err := newTestClient(r.T(), logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()

	r.repository, err = postgres.NewLineageRepository(client)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *LineageRepositoryTestSuite) TestGetGraph() {
	rootNode := "test-get-graph-root-node"

	// populate root node
	// Graph:
	//
	// table-50																							  metabase-tgg-51
	//  				> optimus-tgg-1 >	rootNode > metabase-tgg-99 >
	// table-51 																							metabase-tgg-52
	//
	err := r.repository.Upsert(r.ctx, rootNode, []string{"optimus-tgg-1"}, []string{"metabase-tgg-99"})
	r.Require().NoError(err)
	// populate upstream's node
	err = r.repository.Upsert(r.ctx, "optimus-tgg-1", []string{"table-50", "table-51"}, nil)
	r.Require().NoError(err)
	// populate downstream's node
	err = r.repository.Upsert(r.ctx, "metabase-tgg-99", nil, []string{"metabase-tgg-51", "metabase-tgg-52"})
	r.Require().NoError(err)

	r.Run("should recursively fetch all graph", func() {
		expected := asset.LineageGraph{
			{Source: "optimus-tgg-1", Target: rootNode},
			{Source: "table-50", Target: "optimus-tgg-1"},
			{Source: "table-51", Target: "optimus-tgg-1"},
			{Source: rootNode, Target: "metabase-tgg-99"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-51"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-52"},
		}

		graph, err := r.repository.GetGraph(r.ctx, rootNode, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(expected, graph)
	})

	r.Run("should fetch based on the level given in config if any", func() {
		expected := asset.LineageGraph{
			{Source: "optimus-tgg-1", Target: rootNode},
			{Source: rootNode, Target: "metabase-tgg-99"},
		}

		graph, err := r.repository.GetGraph(r.ctx, rootNode, asset.LineageQuery{
			Level: 1,
		})
		r.Require().NoError(err)
		r.compareGraphs(expected, graph)
	})

	r.Run("should fetch based on the direction given in config if any", func() {
		expected := asset.LineageGraph{
			{Source: rootNode, Target: "metabase-tgg-99"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-51"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-52"},
		}

		graph, err := r.repository.GetGraph(r.ctx, rootNode, asset.LineageQuery{
			Direction: asset.LineageDirectionDownstream,
		})
		r.Require().NoError(err)
		r.compareGraphs(expected, graph)
	})

	r.Run("should be able control soft deleted assets inclusiveness in lineage with includeDeleted config", func() {
		nodeURN := "table-1"

		// create initial
		err := r.repository.Upsert(r.ctx, nodeURN, []string{"table-2"}, []string{"table-3"})
		r.NoError(err)

		err = r.repository.SoftDeleteByURN(r.ctx, nodeURN)
		r.NoError(err)

		// fetch without include soft deleted result
		graph, err := r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{}, graph)

		// fetch with include soft deleted result
		graph, err = r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{IncludeDeleted: true})
		r.Require().NoError(err)
		r.compareGraphsWithProp(asset.LineageGraph{
			{Source: "table-2", Target: nodeURN, Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": false,
				"target_is_deleted": true,
			}},
			{Source: nodeURN, Target: "table-3", Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": true,
				"target_is_deleted": false,
			}},
		}, graph)
	})
}

func (r *LineageRepositoryTestSuite) TestDeleteByURN() {
	r.Run("should delete asset from lineage", func() {
		nodeURN := "table-1"

		// create initial
		err := r.repository.Upsert(r.ctx, nodeURN, []string{"table-2"}, []string{"table-3"})
		r.NoError(err)

		err = r.repository.DeleteByURN(r.ctx, nodeURN)
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{}, graph)
	})

	r.Run("delete when URN has no lineage", func() {
		nodeURN := "table-1"
		err := r.repository.DeleteByURN(r.ctx, nodeURN)
		r.NoError(err)
	})
}

func (r *LineageRepositoryTestSuite) TestSoftDeleteByURN() {
	r.Run("should soft delete asset from lineage", func() {
		nodeURN := "table-1"

		// create initial
		err := r.repository.Upsert(r.ctx, nodeURN, []string{"table-2"}, []string{"table-3"})
		r.NoError(err)

		err = r.repository.SoftDeleteByURN(r.ctx, nodeURN)
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{IncludeDeleted: true})
		r.Require().NoError(err)
		r.compareGraphsWithProp(asset.LineageGraph{
			{Source: "table-2", Target: nodeURN, Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": false,
				"target_is_deleted": true,
			}},
			{Source: nodeURN, Target: "table-3", Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": true,
				"target_is_deleted": false,
			}},
		}, graph)
	})

	r.Run("delete when URN has no lineage", func() {
		nodeURN := "table-1"
		err := r.repository.SoftDeleteByURN(r.ctx, nodeURN)
		r.NoError(err)
	})
}

func (r *LineageRepositoryTestSuite) TestDeleteByURNs() {
	r.Run("should delete assets from lineage", func() {
		nodeURN1a := "table-1a"
		nodeURN1b := "table-1b"
		nodeURNs := []string{nodeURN1a, nodeURN1b}

		// create initial
		err := r.repository.Upsert(r.ctx, nodeURN1a, []string{"table-2"}, []string{"table-3"})
		r.NoError(err)
		err = r.repository.Upsert(r.ctx, nodeURN1b, []string{"table-2"}, []string{"table-3"})
		r.NoError(err)

		err = r.repository.DeleteByURNs(r.ctx, nodeURNs)
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, nodeURN1a, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{}, graph)

		graph, err = r.repository.GetGraph(r.ctx, nodeURN1b, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{}, graph)
	})

	r.Run("delete when URNs has no lineage", func() {
		nodeURN1a := "table-1a"
		nodeURN1b := "table-1b"
		nodeURNs := []string{nodeURN1a, nodeURN1b}

		err := r.repository.DeleteByURNs(r.ctx, nodeURNs)
		r.NoError(err)
	})
}

func (r *LineageRepositoryTestSuite) TestSoftDeleteByURNs() {
	r.Run("should delete assets from lineage", func() {
		nodeURN1a := "table-1a"
		nodeURN1b := "table-1b"
		nodeURNs := []string{nodeURN1a, nodeURN1b}

		// create initial
		err := r.repository.Upsert(r.ctx, nodeURN1a, []string{"table-2"}, []string{"table-3"})
		r.NoError(err)
		err = r.repository.Upsert(r.ctx, nodeURN1b, []string{"table-2"}, []string{"table-3"})
		r.NoError(err)

		err = r.repository.SoftDeleteByURNs(r.ctx, nodeURNs)
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, nodeURN1a, asset.LineageQuery{IncludeDeleted: true})
		r.Require().NoError(err)
		r.compareGraphsWithProp(asset.LineageGraph{
			{Source: "table-2", Target: nodeURN1a, Prop: map[string]interface{}{
				"root":              nodeURN1a,
				"source_is_deleted": false,
				"target_is_deleted": true,
			}},
			{Source: nodeURN1a, Target: "table-3", Prop: map[string]interface{}{
				"root":              nodeURN1a,
				"source_is_deleted": true,
				"target_is_deleted": false,
			}},
		}, graph)

		graph, err = r.repository.GetGraph(r.ctx, nodeURN1b, asset.LineageQuery{IncludeDeleted: true})
		r.Require().NoError(err)
		r.compareGraphsWithProp(asset.LineageGraph{
			{Source: "table-2", Target: nodeURN1b, Prop: map[string]interface{}{
				"root":              nodeURN1b,
				"source_is_deleted": false,
				"target_is_deleted": true,
			}},
			{Source: nodeURN1b, Target: "table-3", Prop: map[string]interface{}{
				"root":              nodeURN1b,
				"source_is_deleted": true,
				"target_is_deleted": false,
			}},
		}, graph)
	})

	r.Run("not return error when URNs has no lineage", func() {
		nodeURN1a := "table-1a"
		nodeURN1b := "table-1b"
		nodeURNs := []string{nodeURN1a, nodeURN1b}

		err := r.repository.SoftDeleteByURNs(r.ctx, nodeURNs)
		r.NoError(err)
	})
}

func (r *LineageRepositoryTestSuite) TestUpsert() {
	r.Run("should insert all as graph if upstreams and downstreams are new", func() {
		nodeURN := "table-1"
		upstreams := []string{"job-1"}
		downstreams := []string{"dashboard-1", "dashboard-2"}
		err := r.repository.Upsert(r.ctx, nodeURN, upstreams, downstreams)
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{
			{Source: "job-1", Target: nodeURN},
			{Source: nodeURN, Target: "dashboard-1"},
			{Source: nodeURN, Target: "dashboard-2"},
		}, graph)
	})

	r.Run("should insert or delete graph when updating existing graph", func() {
		nodeURN := "update-table"

		// create initial
		err := r.repository.Upsert(r.ctx, nodeURN, []string{"job-99"}, []string{"dashboard-99"})
		r.NoError(err)

		// update
		err = r.repository.Upsert(r.ctx, nodeURN, []string{"job-99", "job-100"}, []string{"dashboard-93"})
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{
			{Source: "job-99", Target: nodeURN},
			{Source: "job-100", Target: nodeURN},
			{Source: nodeURN, Target: "dashboard-93"},
		}, graph)
	})

	r.Run("should restore soft deleted edge when re-sync graph", func() {
		nodeURN := "restore-table"

		// create initial
		err := r.repository.Upsert(r.ctx, nodeURN, []string{"job-restore"}, []string{"dashboard-restore"})
		r.NoError(err)
		graph, err := r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{IncludeDeleted: true})
		r.Require().NoError(err)
		r.compareGraphsWithProp(asset.LineageGraph{
			{Source: "job-restore", Target: nodeURN, Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": false,
				"target_is_deleted": false,
			}},
			{Source: nodeURN, Target: "dashboard-restore", Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": false,
				"target_is_deleted": false,
			}},
		}, graph)

		// soft delete
		err = r.repository.SoftDeleteByURN(r.ctx, nodeURN)
		r.NoError(err)
		graph, err = r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{IncludeDeleted: true})
		r.Require().NoError(err)
		r.compareGraphsWithProp(asset.LineageGraph{
			{Source: "job-restore", Target: nodeURN, Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": false,
				"target_is_deleted": true,
			}},
			{Source: nodeURN, Target: "dashboard-restore", Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": true,
				"target_is_deleted": false,
			}},
		}, graph)

		// re-sync
		err = r.repository.Upsert(r.ctx, nodeURN, []string{"job-restore"}, []string{"dashboard-restore"})
		r.NoError(err)
		graph, err = r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{IncludeDeleted: true})
		r.Require().NoError(err)
		r.compareGraphsWithProp(asset.LineageGraph{
			{Source: "job-restore", Target: nodeURN, Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": false,
				"target_is_deleted": false,
			}},
			{Source: nodeURN, Target: "dashboard-restore", Prop: map[string]interface{}{
				"root":              nodeURN,
				"source_is_deleted": false,
				"target_is_deleted": false,
			}},
		}, graph)
	})
}

func (r *LineageRepositoryTestSuite) compareGraphs(expected, actual asset.LineageGraph) {
	expLen := len(expected)
	r.Require().Len(actual, expLen)

	for i := 0; i < expLen; i++ {
		r.Equal(expected[i].Source, actual[i].Source, fmt.Sprintf("different source on index %d", i))
		r.Equal(expected[i].Target, actual[i].Target, fmt.Sprintf("different target on index %d", i))
	}
}

func (r *LineageRepositoryTestSuite) compareGraphsWithProp(expected, actual asset.LineageGraph) {
	expLen := len(expected)
	r.Require().Len(actual, expLen)

	for i := 0; i < expLen; i++ {
		r.Equal(expected[i].Source, actual[i].Source, fmt.Sprintf("different source on index %d", i))
		r.Equal(expected[i].Target, actual[i].Target, fmt.Sprintf("different target on index %d", i))
		r.Equal(expected[i].Prop, actual[i].Prop, fmt.Sprintf("different prop on index %d", i))
	}
}

func TestLineageRepository(t *testing.T) {
	suite.Run(t, &LineageRepositoryTestSuite{})
}
