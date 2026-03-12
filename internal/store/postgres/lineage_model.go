package postgres

import (
	"github.com/goto/compass/core/asset"
)

type LineageGraphModel []LineageEdgeModel

func (gm LineageGraphModel) toGraph() asset.LineageGraph {
	graph := asset.LineageGraph{}
	for _, em := range gm {
		graph = append(graph, em.toEdge())
	}

	return graph
}

type LineageEdgeModel struct {
	Source       string  `db:"source"`
	SourceColumn string  `db:"source_column"`
	Target       string  `db:"target"`
	TargetColumn string  `db:"target_column"`
	Prop         JSONMap `db:"prop"`
}

func (m LineageEdgeModel) toEdge() asset.LineageEdge {
	edge := asset.LineageEdge{
		Source: m.Source,
		Target: m.Target,
		Prop:   m.Prop,
	}

	if m.SourceColumn != "" {
		edge.SourceColumn = m.SourceColumn
	}

	if m.TargetColumn != "" {
		edge.TargetColumn = m.TargetColumn
	}

	return edge
}
