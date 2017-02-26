// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder parses build graphs and coordinates builds
package builder

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"sync"

	"bldy.build/build"
	"bldy.build/build/blaze"
	"bldy.build/build/blaze/postprocessor"
	"bldy.build/build/project"
	"bldy.build/build/racy"
	bldytrg "bldy.build/build/targets/build"
	"bldy.build/build/url"
)

type Update struct {
	TimeStamp time.Time
	Target    string
	Status    STATUS
	Worker    int
	Cached    bool
}

type Builder struct {
	Origin      string
	Wd          string
	ProjectPath string
	Nodes       map[string]*Node
	Total       int
	Done        chan *Node
	Error       chan error
	Timeout     chan bool
	Updates     chan *Node
	Root, ptr   *Node
	pq          *p
	vm          *blaze.VM
}

func New() (c Builder) {
	c.Nodes = make(map[string]*Node)
	c.Error = make(chan error)
	c.Done = make(chan *Node)
	c.Updates = make(chan *Node)
	var err error
	c.Wd, err = os.Getwd()
	if err != nil {
		l.Fatal(err)
	}
	c.pq = newP()
	c.ProjectPath = project.Root()
	c.vm = blaze.NewVM(c.Wd)
	return
}

type Node struct {
	IsRoot     bool         `json:"-"`
	Target     build.Target `json:"-"`
	Type       string
	Parents    map[string]*Node `json:"-"`
	Url        url.URL
	Worker     string
	Priority   int
	wg         sync.WaitGroup
	Status     STATUS
	Cached     bool
	Start, End int64
	Hash       string
	Output     string `json:"-"`
	once       sync.Once
	sync.Mutex
	Children map[string]*Node
	hash     []byte
}

func (n *Node) priority() int {
	if n.Priority < 0 {
		p := 0
		for _, c := range n.Parents {
			p += c.priority() + 1
		}
		n.Priority = p
	}
	return n.Priority
}
func (b *Builder) getTarget(u url.URL) (n *Node) {
	if gnode, ok := b.Nodes[u.String()]; ok {
		return gnode
	}
	t, err := b.vm.GetTarget(u)
	if err != nil {
		log.Fatal(err)
	}
	xu := url.URL{
		Package: u.Package,
		Target:  t.GetName(),
	}

	node := Node{
		Target:   t,
		Type:     fmt.Sprintf("%T", t)[1:],
		Children: make(map[string]*Node),
		Parents:  make(map[string]*Node),
		once:     sync.Once{},
		wg:       sync.WaitGroup{},
		Status:   Pending,
		Url:      xu,
		Priority: -1,
	}

	post := postprocessor.New(u.Package)

	err = post.ProcessDependencies(node.Target)
	if err != nil {
		l.Fatal(err)
	}

	var deps []build.Target

	//group is a special case
	var group *bldytrg.Group
	switch node.Target.(type) {
	case *bldytrg.Group:
		group = node.Target.(*bldytrg.Group)
		group.Exports = make(map[string]string)
	}
	for _, d := range node.Target.GetDependencies() {
		c := b.Add(d)
		node.wg.Add(1)
		if group != nil {
			for dst, _ := range c.Target.Installs() {
				group.Exports[dst] = dst
			}
		}
		deps = append(deps, c.Target)

		node.Children[d] = c
		c.Parents[xu.String()] = &node
	}

	if err := post.ProcessPaths(t, deps); err != nil {
		l.Fatalf("path processing: %s", err.Error())
	}

	b.Nodes[xu.String()] = &node
	if t.GetName() == u.Target {
		n = &node
	} else {
		l.Fatalf("target name %q and url target %q don't match", t.GetName(), u.Target)
	}
	return n
}

func (b *Builder) Add(t string) *Node {
	return b.getTarget(url.Parse(t))
}

func (n *Node) HashNode() []byte {

	// node hashes should not change after a build,
	// they should be deterministic, therefore they can and should be cached.
	if len(n.hash) > 0 {
		return n.hash
	}
	n.hash = n.Target.Hash()
	var bn ByName
	for _, e := range n.Children {
		bn = append(bn, e)
	}
	sort.Sort(bn)
	for _, e := range bn {
		n.hash = racy.XOR(e.HashNode(), n.hash)
	}
	n.Hash = fmt.Sprintf("%x", n.hash)
	return n.hash
}

type ByName []*Node

func (a ByName) Len() int      { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool {
	return strings.Compare(a[i].Target.GetName(), a[j].Target.GetName()) > 0
}
