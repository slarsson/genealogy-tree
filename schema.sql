DROP DATABASE IF EXISTS genealogy;

CREATE DATABASE genealogy;

CREATE TABLE IF NOT EXISTS edge (
  source_node_id   VARCHAR(64),
  source_node_type VARCHAR(64),
  target_node_id   VARCHAR(64),
  target_node_type VARCHAR(64),
  PRIMARY KEY (source_node_id, target_node_id)
);

CREATE INDEX adjacent_parent_idx ON edge (target_node_id, source_node_id);
