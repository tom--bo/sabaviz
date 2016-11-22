package main

import (
	// "fmt"
	"github.com/awalterschulze/gographviz"
	// "io"
	// "strings"
)

type Edge struct {
	host1 string
	host2 string
	port  string
}

type Graph struct {
	graph   *gographviz.Graph
	edgeMap map[Edge]bool
}

func (g *Graph) NewGraph() {
	graphAst, _ := gographviz.Parse([]byte(`digraph G{}`))
	g.graph = gographviz.NewGraph()
	gographviz.Analyse(graphAst, g.graph)
	g.edgeMap = make(map[Edge]bool)
}

func (g *Graph) AddNode(h Host) {
	g.graph.AddNode("G", h.hostName, nil)
}

func (g *Graph) AddConnectionOnce(h Host, conn Connection) {
	var edge Edge
	if h.hostName < conn.hostName {
		edge = Edge{h.hostName, conn.hostName, conn.port}
	} else {
		edge = Edge{conn.hostName, h.hostName, conn.port}
	}
	_, ok := g.edgeMap[edge]
	if !ok {
		g.edgeMap[edge] = true
		g.graph.AddEdge(h.hostName, conn.hostName, false, map[string]string{"label": conn.port})
	}
}
