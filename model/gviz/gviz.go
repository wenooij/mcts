package gviz

import (
	"bytes"
	"fmt"
	"io"

	"github.com/wenooij/mcts"
)

type Tree struct {
	elem     mcts.Node
	children map[mcts.Action]*Tree
	pv       bool
}

func (t *Tree) Add(v mcts.Variation, pv bool) {
	if len(v) == 0 {
		return
	}
	if t.elem.Action == nil {
		t.elem = v[0]
	}
	if len(v) == 1 {
		return
	}
	for _, e := range v.TrimRoot() {
		if pv {
			t.pv = pv
		}
		if t.children == nil {
			t.children = make(map[mcts.Action]*Tree)
		}
		subtree := t.children[e.Action()]
		if subtree == nil {
			subtree = new(Tree)
			subtree.elem = e
			t.children[e.Action()] = subtree
		}
		t = subtree
	}
}

func (t *Tree) DOT(w io.Writer) (n int, err error) {
	return (&treeDotter{Tree: t}).DOT(w)
}

type treeDotter struct {
	nextId int
	*Tree
}

func (t *treeDotter) DOT(w io.Writer) (n int, err error) {
	var b bytes.Buffer
	b.WriteString("digraph G {\n")
	defer func() {
		b.WriteString("}\n")
		n, err = w.Write(b.Bytes())
	}()
	parent := t.writeNode(&b, t.elem, t.pv)
	t.recDOT(&b, parent)
	return
}

func (t *treeDotter) recDOT(b *bytes.Buffer, parent string) {
	for _, child := range t.children {
		a := t.writeNode(b, child.elem, child.pv)
		fmt.Fprintf(b, "  %s -> %s;\n", parent, a)
		tree := t.Tree
		t.Tree = child
		t.recDOT(b, a)
		t.Tree = tree
	}
}

func (t *treeDotter) writeNode(b *bytes.Buffer, s mcts.Node, pv bool) string {
	id := fmt.Sprint(t.nextId)
	actionStr := "<root>"
	if s.Action != nil {
		actionStr = s.Action().String()
	}
	nodeTypeStr := ""
	pvStyle := ""
	if pv {
		nodeTypeStr = " PV"
		pvStyle = ` style=filled, color="#A4FD78",`
	}
	if s.Terminal() {
		nodeTypeStr += " #"
	}
	t.nextId++
	fmt.Fprintf(b, `%s [shape=square,%s label="%s%s\lrollouts: %.0f\lscore: %.2f\lpriority: %.2f\l"];`,
		id, pvStyle, actionStr, nodeTypeStr, s.NumRollouts(), s.Score(), s.Priority())
	b.WriteByte('\n')
	return id
}

func VarDOT(w io.Writer, v mcts.Variation) (n int, err error) {
	var t Tree
	t.Add(v, true)
	return t.DOT(w)
}
