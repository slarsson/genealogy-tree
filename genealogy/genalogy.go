package genealogy

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Service interface {
	// Get immediate descendants/children
	Children(ctx context.Context, nodeID string) ([]Node, error)

	// Get all descendants
	Descendants(ctx context.Context, nodeID string) ([]Node, error)

	// Get first descendants of type
	FirstDescendantsOfType(ctx context.Context, nodeID string, nodeType string) ([]Node, error)

	// Get immediate ascendants/parents
	Parents(ctx context.Context, nodeID string) ([]Node, error)

	// Get all ascendants
	Ascendants(ctx context.Context, nodeID string) ([]Node, error)

	// Adds an edge between two nodes
	AddEdge(ctx context.Context, source Node, target Node) error

	// Removes an edge between two nodes
	RemoveEdge(ctx context.Context, source Node, target Node) error
}

var _ Service = &Genealogy{}

type Genealogy struct {
	tableName string
	db        *sql.DB
}

type Node struct {
	ID   string
	Type string
}

func New(pgConnectString string, tableName string) (*Genealogy, error) {
	db, err := sql.Open("postgres", pgConnectString)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	return &Genealogy{
		db:        db,
		tableName: tableName,
	}, nil
}

func (g *Genealogy) AddEdge(ctx context.Context, source Node, target Node) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (source_node_id, source_node_type, target_node_id, target_node_type) 
		VALUES ($1, $2, $3, $4)
	`, g.tableName)

	stmt, err := g.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, source.ID, source.Type, target.ID, target.Type)
	return err
}

func (g *Genealogy) RemoveEdge(ctx context.Context, source Node, target Node) error {
	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE source_node_id = $1 AND source_node_type = $2 AND target_node_id = $3 AND target_node_type = $4
	`, g.tableName)

	stmt, err := g.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, source.ID, source.Type, target.ID, target.Type)
	return err
}

func (g *Genealogy) Children(ctx context.Context, nodeID string) ([]Node, error) {
	query := fmt.Sprintf(`
		SELECT target_node_id, target_node_type FROM %s
		WHERE source_node_id = $1
	`, g.tableName)

	return g.queryNodes(ctx, query, nodeID)
}

func (g *Genealogy) Descendants(ctx context.Context, nodeID string) ([]Node, error) {
	query := fmt.Sprintf(`
		WITH RECURSIVE descendants AS (
			SELECT target_node_id AS node_id, target_node_type FROM %s
			WHERE source_node_id = $1
			
			UNION ALL
		
			SELECT e.target_node_id, e.target_node_type
			FROM descendants
			JOIN %s AS e ON e.source_node_id = descendants.node_id
		)
		SELECT DISTINCT node_id, target_node_type FROM descendants
	`, g.tableName, g.tableName)

	return g.queryNodes(ctx, query, nodeID)
}

func (g *Genealogy) FirstDescendantsOfType(ctx context.Context, nodeID string, nodeType string) ([]Node, error) {
	query := fmt.Sprintf(`
		WITH RECURSIVE descendants AS (
			SELECT target_node_id AS node_id, target_node_type FROM %s
			WHERE source_node_id = $1
			
			UNION ALL
		
			SELECT e.target_node_id, e.target_node_type
			FROM descendants
			JOIN %s AS e ON e.source_node_id = descendants.node_id
			WHERE NOT e.source_node_type = $2
		)
		SELECT DISTINCT node_id, target_node_type FROM descendants
		WHERE target_node_type = $2
	`, g.tableName, g.tableName)

	return g.queryNodes(ctx, query, nodeID, nodeType)
}

func (g *Genealogy) Parents(ctx context.Context, nodeID string) ([]Node, error) {
	query := fmt.Sprintf(`
		SELECT source_node_id, source_node_type FROM %s
		WHERE target_node_id = $1
	`, g.tableName)

	return g.queryNodes(ctx, query, nodeID)
}

func (g *Genealogy) Ascendants(ctx context.Context, nodeID string) ([]Node, error) {
	// TODO: how to ignore duplicates during recursive?
	query := fmt.Sprintf(`
		WITH RECURSIVE ascendants AS (
			SELECT source_node_id AS node_id, source_node_type FROM %s
			WHERE target_node_id = $1
			
			UNION ALL
		
			SELECT e.source_node_id, e.source_node_type
			FROM ascendants
			JOIN %s AS e ON e.target_node_id = ascendants.node_id
		)
		SELECT DISTINCT node_id, source_node_type FROM ascendants
	`, g.tableName, g.tableName)

	return g.queryNodes(ctx, query, nodeID)
}

func (g *Genealogy) queryNodes(ctx context.Context, query string, args ...any) ([]Node, error) {
	stmt, err := g.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}

	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("queryContext: %w", err)
	}

	var nodes []Node

	defer rows.Close()
	for rows.Next() {
		var id, typ string
		if err := rows.Scan(&id, &typ); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		nodes = append(nodes, Node{
			ID:   id,
			Type: typ,
		})
	}

	return nodes, nil
}
