package gviz

import (
	"bytes"
	"fmt"
	"io"

	"github.com/wenooij/heapordered"
	"github.com/wenooij/mcts"
)

func DOT[T mcts.Counter](w io.Writer, t *heapordered.Tree[mcts.Node[T]]) (n int, err error) {
	var b bytes.Buffer
	b.WriteString("digraph G {\n")
	parent := writeNode(&b, t, 0, true, true)
	recDOT(&b, parent, t, 1)
	b.WriteString("}\n")
	n, err = w.Write(b.Bytes())
	return
}

func recDOT[T mcts.Counter](b *bytes.Buffer, parent string, t *heapordered.Tree[mcts.Node[T]], nextId int) int {
	for i := 0; i < t.Len(); i++ {
		child := t.At(i)
		a := writeNode(b, child, nextId, false, false)
		nextId++
		fmt.Fprintf(b, "  %s -> %s;\n", parent, a)
		nextId = recDOT[T](b, a, child, nextId)
	}
	return nextId
}

func writeNode[T mcts.Counter](b *bytes.Buffer, s *heapordered.Tree[mcts.Node[T]], nextId int, root, pv bool) string {
	id := fmt.Sprint(nextId)
	actionStr := "<root>"
	e := s.E
	if e.Action != nil {
		actionStr = e.Action.String()
	}
	nodeTypeStr := ""
	pvStyle := ""
	if pv {
		nodeTypeStr = " PV"
		pvStyle = ` style=filled, color="#A4FD78",`
	}
	if s.Len() == 0 {
		nodeTypeStr += " #"
	}
	fmt.Fprintf(b, `%s [shape=square,%s label="%s%s\lrollouts: %f\lscore: %.2f\l"];`,
		id, pvStyle, actionStr, nodeTypeStr, e.NumRollouts, e.Score.Apply())
	b.WriteByte('\n')
	return id
}
