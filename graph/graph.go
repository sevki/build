// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package graph parses and generates build graphs
package graph

import (
	"log"
	"os"

	"bldy.build/build/url"

	"bldy.build/build"
)

var (
	l = log.New(os.Stdout, "graph: ", 0)
)

// New returns a new build graph relatvie to the working directory
func New(u *url.URL) (*Graph, error) {
	g := Graph{
		Nodes: make(map[string]*Node),
	}
	g.Root = g.getTarget(u)
	g.Root.IsRoot = true
	return &g, nil
}

// Graph represents a build graph
type Graph struct {
	Root      *Node
	vm        build.VM
	Workspace string
	Nodes     map[string]*Node
}

func (g *Graph) getTarget(u *url.URL) (n *Node) {
	if gnode, ok := g.Nodes[u.String()]; ok {
		return gnode
	}

	t, err := g.vm.GetTarget(u)
	if err != nil {
		l.Fatal(err)
	}

	node := NewNode(u, t)

	var deps []build.Rule

	for _, d := range node.Target.Dependencies() {
		c := g.getTarget(d)
		if err != nil {
			l.Printf("%q is not a valid label", d.String())
			continue
		}
		node.WG.Add(1)

		deps = append(deps, c.Target)

		node.Children[d.String()] = c
		c.Parents[u.String()] = &node
	}

	g.Nodes[u.String()] = &node
	if t.Name() == u.String() {
		n = &node
	} else {
		l.Fatalf("target name %q and url target %q don't match", t.Name(), u.String())
	}
	return n
}
