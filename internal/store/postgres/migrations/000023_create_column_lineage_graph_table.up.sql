CREATE TABLE column_lineage_graph (
    source_asset text NOT NULL,
    source_column text NOT NULL,
    target_asset text NOT NULL,
    target_column text NOT NULL,
    prop jsonb,
    primary key (source_asset, source_column, target_asset, target_column)
);
