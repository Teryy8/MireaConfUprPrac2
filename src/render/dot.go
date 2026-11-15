// Package render переводит построенный граф в Graphviz DOT и SVG.
package render

import (
	"fmt"
	"mirea/prac2/src/graph"
	"strconv"
	"strings"
)

// DOT возвращает текст Graphviz с направлением от пакета к его зависимостям.
func DOT(g *graph.Graph) string {
	var out strings.Builder
	out.WriteString("digraph Dependencies {\n  rankdir=LR;\n")
	out.WriteString("  graph [bgcolor=white];\n  node [shape=box, fontname=Arial, style=filled, fillcolor=whitesmoke];\n")
	if g.Root == "" {
		out.WriteString("  label=\"Пакеты исключены фильтром\";\n")
	}
	for _, id := range g.Order {
		fmt.Fprintf(&out, "  %s [%s];\n", strconv.Quote(id), nodeAttributes(g, id))
	}
	cycles := make(map[graph.Edge]bool)
	for _, edge := range g.Cycles {
		cycles[edge] = true
	}
	for _, from := range g.Order {
		for _, to := range g.Edges[from] {
			style := ""
			if cycles[graph.Edge{From: from, To: to}] {
				style = " [color=red, penwidth=2]"
			}
			fmt.Fprintf(&out, "  %s -> %s%s;\n", strconv.Quote(from), strconv.Quote(to), style)
		}
	}
	out.WriteString("}\n")
	return out.String()
}

// nodeAttributes выделяет корень и помечает нераскрытые узлы на границе глубины.
func nodeAttributes(g *graph.Graph, id string) string {
	label := id
	attributes := ""
	if id == g.Root {
		attributes = ", fillcolor=lightblue"
	}
	if g.Depth[id] >= g.Limit && len(g.Nodes[id].Dependencies) > 0 && len(g.Edges[id]) == 0 {
		label += "\n(предел глубины)"
		attributes += ", style=\"filled,dashed\""
	}
	return "label=" + strconv.Quote(label) + attributes
}
