CREATE INDEX idx_lineage_graph_target ON lineage_graph (target);
CREATE INDEX idx_lineage_graph_not_deleted ON lineage_graph (source, target) WHERE (prop->>'source_is_deleted') = 'false' AND (prop->>'target_is_deleted') = 'false';
CREATE INDEX idx_column_lineage_target ON column_lineage_graph (target_asset, target_column);
CREATE INDEX idx_column_lineage_not_deleted ON column_lineage_graph (source_asset, source_column, target_asset, target_column) WHERE (prop->>'source_is_deleted') = 'false' AND (prop->>'target_is_deleted') = 'false';
