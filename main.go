package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/slarsson/genealogy-tree/genealogy"
)

func main() {
	const conn = "postgresql://postgres:postgres@0.0.0.0:5432/genealogy?sslmode=disable"
	const tableName = "edge"

	ctx := context.Background()

	var svc genealogy.Service
	var err error

	svc, err = genealogy.New(conn, tableName)
	if err != nil {
		log.Fatal(err)
	}

	c1 := uuid.NewString()
	c2 := uuid.NewString()
	p1 := uuid.NewString()
	p2 := uuid.NewString()

	for _, nodes := range exampleTree(c1, c2, p1, p2) {
		if err := svc.AddEdge(ctx, nodes.source, nodes.target); err != nil {
			log.Fatal(err)
		}
	}

	nodes, err := svc.Descendants(ctx, p1)
	// nodes, err := svc.Ascendants(ctx, c1)
	if err != nil {
		log.Fatal(err)
	}
	for i, node := range nodes {
		fmt.Println(i+1, node.ID, node.Type)
	}
}

type TestData struct {
	source genealogy.Node
	target genealogy.Node
}

func exampleTree(c1, c2, p1, p2 string) []TestData {

	// p1 -> s1..10 -> j1 -> c1
	// p1 -> s11 -> j2 -> c1
	// p1 -> s12 -> j3 -> c2
	// p2 -> s13 -> j3 -> c2

	var out []TestData

	j1 := uuid.NewString()
	j2 := uuid.NewString()
	j3 := uuid.NewString()

	// p1 -> s1..10 -> j1 -> c1
	out = append(out, TestData{
		source: genealogy.Node{ID: j1, Type: "J"},
		target: genealogy.Node{ID: c1, Type: "C"},
	})
	for i := 1; i <= 10; i++ {
		s := uuid.NewString()
		out = append(out, TestData{
			source: genealogy.Node{ID: s, Type: "S"},
			target: genealogy.Node{ID: j1, Type: "J"},
		})
		out = append(out, TestData{
			source: genealogy.Node{ID: p1, Type: "P"},
			target: genealogy.Node{ID: s, Type: "S"},
		})
	}

	// p1 -> s11 -> j2 -> c1
	s11 := uuid.NewString()
	out = append(out, TestData{
		source: genealogy.Node{ID: j2, Type: "J"},
		target: genealogy.Node{ID: c1, Type: "C"},
	})
	out = append(out, TestData{
		source: genealogy.Node{ID: s11, Type: "S"},
		target: genealogy.Node{ID: j2, Type: "J"},
	})
	out = append(out, TestData{
		source: genealogy.Node{ID: p1, Type: "P"},
		target: genealogy.Node{ID: s11, Type: "S"},
	})

	// p1 -> s12 -> j3 -> c2
	s12 := uuid.NewString()
	out = append(out, TestData{
		source: genealogy.Node{ID: j3, Type: "J"},
		target: genealogy.Node{ID: c2, Type: "C"},
	})
	out = append(out, TestData{
		source: genealogy.Node{ID: s12, Type: "S"},
		target: genealogy.Node{ID: j3, Type: "J"},
	})
	out = append(out, TestData{
		source: genealogy.Node{ID: p1, Type: "P"},
		target: genealogy.Node{ID: s12, Type: "S"},
	})

	// p2 -> s13 -> j3 -> c2
	s13 := uuid.NewString()
	out = append(out, TestData{
		source: genealogy.Node{ID: s13, Type: "S"},
		target: genealogy.Node{ID: j3, Type: "J"},
	})
	out = append(out, TestData{
		source: genealogy.Node{ID: p2, Type: "P"},
		target: genealogy.Node{ID: s13, Type: "S"},
	})

	return out
}
