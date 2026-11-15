// Текстовый список рёбер для проверки построенного графа.
package graph

import (
	"fmt"
	"io"
	"strings"
)

// Print показывает узлы в порядке DFS, рёбра и обратные рёбра обнаруженных циклов.
func (g *Graph) Print(w io.Writer) error {
	if g.Root == "" {
		_, err := fmt.Fprintln(w, "Корневой пакет исключён фильтром.")
		return err
	}
	if _, err := fmt.Fprintln(w, "Граф зависимостей:"); err != nil {
		return err
	}
	for _, id := range g.Order {
		targets := strings.Join(g.Edges[id], ", ")
		if targets == "" {
			targets = g.leafLabel(id)
		}
		if _, err := fmt.Fprintf(w, "%s -> %s\n", id, targets); err != nil {
			return err
		}
	}
	for _, edge := range g.Cycles {
		if _, err := fmt.Fprintf(w, "Цикл: %s -> %s\n", edge.From, edge.To); err != nil {
			return err
		}
	}
	return nil
}

// leafLabel различает лист, фильтрацию и остановку на границе глубины.
func (g *Graph) leafLabel(id string) string {
	if len(g.Nodes[id].Dependencies) == 0 {
		return "(нет зависимостей)"
	}
	if g.Depth[id] >= g.Limit {
		return "(предел глубины)"
	}
	return "(зависимости исключены фильтром)"
}
