package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/goto/compass/core/asset"
	"github.com/jmoiron/sqlx"
)

type LineageRepository struct {
	client *Client
}

// NewLineageRepository initializes lineage repository
func NewLineageRepository(client *Client) (*LineageRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &LineageRepository{
		client: client,
	}, nil
}

// GetGraph returns a graph that contains list of relations of a given node
func (repo *LineageRepository) GetGraph(ctx context.Context, urn string, query asset.LineageQuery) (asset.LineageGraph, error) {
	var graph asset.LineageGraph

	if query.Direction == "" || query.Direction == asset.LineageDirectionUpstream {
		upstreams, err := repo.getUpstreamsGraph(ctx, urn, query.Level)
		if err != nil {
			return graph, fmt.Errorf("error fetching upstreams graph: %w", err)
		}
		graph = append(graph, upstreams...)
	}

	if query.Direction == "" || query.Direction == asset.LineageDirectionDownstream {
		downstreams, err := repo.getDownstreamsGraph(ctx, urn, query.Level)
		if err != nil {
			return graph, fmt.Errorf("error fetching upstreams graph: %w", err)
		}
		graph = append(graph, downstreams...)
	}

	return graph, nil
}

func (repo *LineageRepository) DeleteByURN(ctx context.Context, urn string) error {
	return repo.DeleteByURNs(ctx, []string{urn})
}

func (repo *LineageRepository) DeleteByURNs(ctx context.Context, urns []string) error {
	qry, args, err := sq.Delete("lineage_graph").
		Where(sq.Or{
			sq.Eq{"source": urns},
			sq.Eq{"target": urns},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete query: URN = '%v': %w", urns, err)
	}

	if _, err := repo.client.db.ExecContext(ctx, qry, args...); err != nil {
		return fmt.Errorf("delete asset: URN = '%v': %w", urns, err)
	}

	return nil
}

// Upsert insert or delete connections of a given node by comparing them with current state
func (repo *LineageRepository) Upsert(ctx context.Context, urn string, upstreams, downstreams []string) error {
	currentGraph, err := repo.getDirectLineage(ctx, urn)
	if err != nil {
		return fmt.Errorf("error getting node's direct lineage: %w", err)
	}
	newGraph := repo.buildGraph(urn, upstreams, downstreams)
	toInserts, toRemoves := repo.compareGraph(currentGraph, newGraph)
	toRemoves = repo.filterSelfDeleteOnly(urn, toRemoves)

	return repo.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		err := repo.insertGraph(ctx, tx, toInserts)
		if err != nil {
			return fmt.Errorf("error inserting graph: %w", err)
		}

		err = repo.removeGraph(ctx, tx, toRemoves)
		if err != nil {
			return fmt.Errorf("error removing graph: %w", err)
		}

		return nil
	})
}

func (*LineageRepository) buildGraph(urn string, upstreams, downstreams []string) asset.LineageGraph {
	graph := make(asset.LineageGraph, 0, len(upstreams)+len(downstreams))
	for _, us := range upstreams {
		graph = append(graph, asset.LineageEdge{
			Source: us,
			Target: urn,
			Prop: map[string]interface{}{
				"root": urn, // this is to note which node is updating the relation
			},
		})
	}
	for _, ds := range downstreams {
		graph = append(graph, asset.LineageEdge{
			Source: urn,
			Target: ds,
			Prop: map[string]interface{}{
				"root": urn, // this is to note which node is updating the relation
			},
		})
	}

	return graph
}

// filterSelfDeleteOnly filters edges that are not created by the given node
// it uses prop["root"] field to figure out which node (source or target) is latest updater of the edge,
// and only allow that node to delete the relation
func (*LineageRepository) filterSelfDeleteOnly(urn string, toRemoves asset.LineageGraph) asset.LineageGraph {
	var res asset.LineageGraph
	for _, edge := range toRemoves {
		rootURN, ok := edge.Prop["root"]
		if ok && rootURN != urn {
			continue
		}
		res = append(res, edge)
	}

	return res
}

func (*LineageRepository) insertGraph(ctx context.Context, execer sqlx.ExecerContext, graph asset.LineageGraph) error {
	if len(graph) == 0 {
		return nil
	}

	builder := sq.Insert("lineage_graph").Columns("source", "target", "prop").PlaceholderFormat(sq.Dollar)
	for _, edge := range graph {
		builder = builder.Values(edge.Source, edge.Target, edge.Prop)
	}
	builder = builder.Suffix("ON CONFLICT DO NOTHING")

	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	_, err = execer.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error executing insert query: %w", err)
	}

	return nil
}

func (*LineageRepository) removeGraph(ctx context.Context, execer sqlx.ExecerContext, graph asset.LineageGraph) error {
	if len(graph) == 0 {
		return nil
	}

	builder := sq.Delete("lineage_graph").PlaceholderFormat(sq.Dollar)
	conditions := sq.Or{}
	for _, edge := range graph {
		conditions = append(conditions,
			sq.Eq{"source": edge.Source, "target": edge.Target},
		)
	}
	builder = builder.Where(conditions)

	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}
	_, err = execer.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil
}

func (*LineageRepository) compareGraph(current, new asset.LineageGraph) (toInserts, toRemoves asset.LineageGraph) {
	if len(current) == 0 && len(new) == 0 {
		return nil, nil
	}

	currMap := map[string]asset.LineageEdge{}
	for _, c := range current {
		key := c.Source + c.Target
		currMap[key] = c
	}

	for _, n := range new {
		key := n.Source + n.Target
		_, exists := currMap[key]
		if exists {
			// if exists, it means that both new and current have it.
			// we remove it from the map,
			// so that what's left in the map is the that only exists in current
			// and have to be removed
			delete(currMap, key)
		} else {
			toInserts = append(toInserts, n)
		}
	}

	for _, edge := range currMap {
		toRemoves = append(toRemoves, edge)
	}

	return toInserts, toRemoves
}

func (repo *LineageRepository) getUpstreamsGraph(ctx context.Context, urn string, level int) (asset.LineageGraph, error) {
	var graph asset.LineageGraph

	query, args, err := repo.buildUpstreamQuery(urn, level)
	if err != nil {
		return graph, fmt.Errorf("error building upstream query: %w", err)
	}

	var gm LineageGraphModel
	err = repo.client.db.SelectContext(ctx, &gm, query, args...)
	if err != nil {
		return graph, fmt.Errorf("error running upstream query: %w", err)
	}

	graph = gm.toGraph()

	return graph, nil
}

func (repo *LineageRepository) getDownstreamsGraph(ctx context.Context, urn string, level int) (asset.LineageGraph, error) {
	var graph asset.LineageGraph

	query, args, err := repo.buildDownstreamQuery(urn, level)
	if err != nil {
		return graph, fmt.Errorf("error building downstream query: %w", err)
	}

	var gm LineageGraphModel
	err = repo.client.db.SelectContext(ctx, &gm, query, args...)
	if err != nil {
		return graph, fmt.Errorf("error running downstream query: %w", err)
	}

	graph = gm.toGraph()

	return graph, nil
}

func (repo *LineageRepository) buildUpstreamQuery(urn string, level int) (query string, args []interface{}, err error) {
	alias := "search_graph"
	nonRecursiveBuilder := sq.
		Select("source", "target", "prop", "1 as depth", "ARRAY[target] as path").
		From("lineage_graph").
		Where("target = ?", urn)
	recursiveBuilder := sq.
		Select("lg.source", "lg.target", "lg.prop", "sg.depth + 1", "sg.path || lg.target").
		From(fmt.Sprintf("lineage_graph lg, %s sg", alias)).
		Where("lg.target <> ALL(sg.path)").
		Where("lg.target = sg.source")
	if level > 0 {
		recursiveBuilder = recursiveBuilder.Where("sg.depth < ?", level)
	}

	return repo.buildRecursiveQuery(alias, nonRecursiveBuilder, recursiveBuilder)
}

func (repo *LineageRepository) buildDownstreamQuery(urn string, level int) (query string, args []interface{}, err error) {
	alias := "search_graph"
	nonRecursiveBuilder := sq.
		Select("source", "target", "prop", "1 as depth", "ARRAY[source] as path").
		From("lineage_graph").
		Where("source = ?", urn)
	recursiveBuilder := sq.
		Select("lg.source", "lg.target", "lg.prop", "sg.depth + 1", "sg.path || lg.source").
		From(fmt.Sprintf("lineage_graph lg, %s sg", alias)).
		Where("lg.source <> ALL(sg.path)").
		Where("lg.source = sg.target")
	if level > 0 {
		recursiveBuilder = recursiveBuilder.Where("sg.depth < ?", level)
	}

	return repo.buildRecursiveQuery(alias, nonRecursiveBuilder, recursiveBuilder)
}

func (*LineageRepository) buildRecursiveQuery(alias string, nonRecursiveBuilder, recursiveBuilder sq.SelectBuilder) (
	query string, args []interface{}, err error,
) {
	cteBuilder := recursiveCTEBuilder{
		alias:               alias,
		columns:             []string{"source", "target", "prop", "depth", "path"},
		nonRecursiveBuilder: nonRecursiveBuilder,
		recursiveBuilder:    recursiveBuilder,
	}
	cteQuery, cteArgs, err := cteBuilder.toSQL()
	if err != nil {
		return "", nil, fmt.Errorf("build recursive cte: %w", err)
	}

	query, args, err = sq.
		Select("source", "target", "prop").
		From(alias).
		Prefix(cteQuery).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("build final recursive query: %w", err)
	}

	args = append(cteArgs, args...)
	return query, args, nil
}

func (repo *LineageRepository) getDirectLineage(ctx context.Context, urn string) (asset.LineageGraph, error) {
	query := `SELECT * FROM lineage_graph WHERE (source = $1 OR target = $1)`
	var gm LineageGraphModel
	if err := repo.client.db.SelectContext(ctx, &gm, query, urn); err != nil {
		return nil, fmt.Errorf("run query to fetch direct nodes: %w", err)
	}

	return gm.toGraph(), nil
}

type recursiveCTEBuilder struct {
	alias               string
	columns             []string
	nonRecursiveBuilder sq.SelectBuilder
	recursiveBuilder    sq.SelectBuilder
}

func (b *recursiveCTEBuilder) toSQL() (string, []interface{}, error) {
	query, args, err := b.nonRecursiveBuilder.
		Suffix("UNION ALL").
		SuffixExpr(b.recursiveBuilder).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return "", nil, err
	}

	cols := strings.Join(b.columns, ", ")
	query = fmt.Sprintf(`
		WITH RECURSIVE %s (
			%s
		) AS (%s)`,
		b.alias, cols, query)

	return query, args, nil
}
