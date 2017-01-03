package main

import (
	"github.com/awalterschulze/gographviz"
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
	graphAst, _ := gographviz.Parse([]byte(`graph G{}`))
	g.graph = gographviz.NewGraph()
	gographviz.Analyse(graphAst, g.graph)
	g.edgeMap = make(map[Edge]bool)
}

func (g *Graph) AddNode(hostName string) {
	host := "\"" + hostName + "\""
	g.graph.AddNode("G", host, nil)
}

func (g *Graph) AddConnectionOnce(ori string, conn Connection) {
	var edge Edge
	h1 := "\"" + ori + "\""
	h2 := "\"" + conn.hostName + "\""
	if h1 < h2 {
		edge = Edge{h1, h2, conn.port}
	} else {
		edge = Edge{h2, h1, conn.port}
	}
	_, ok := g.edgeMap[edge]
	if !ok {
		g.edgeMap[edge] = true
        port := "\"" + conn.port + "\""
		g.graph.AddEdge(h1, h2, false, map[string]string{"label": port})
	}
}
