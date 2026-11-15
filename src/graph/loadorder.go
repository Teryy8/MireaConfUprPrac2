// Порядок загрузки: зависимости располагаются раньше использующих их пакетов.
package graph

import (
	"fmt"
	"io"
)

// LoadOrder выполняет топологическую сортировку и отклоняет циклический граф.
func (g *Graph) LoadOrder() ([]string, error) {
	if len(g.Cycles) != 0 {
		return nil, fmt.Errorf("порядок загрузки невозможен: в графе есть цикл")
	}
	visited := make(map[string]bool)
	result := make([]string, 0, len(g.Nodes))
	for _, id := range g.Order {
		g.appendDependencyFirst(id, visited, &result)
	}
	return result, nil
}

// appendDependencyFirst рекурсивно добавляет пакет после его зависимостей.
func (g *Graph) appendDependencyFirst(id string, visited map[string]bool, result *[]string) {
	if visited[id] {
		return
	}
	visited[id] = true
	for _, child := range g.Edges[id] {
		g.appendDependencyFirst(child, visited, result)
	}
	*result = append(*result, id)
}

// PrintLoadOrder выводит пронумерованный порядок для выбранной глубины и фильтра.
func (g *Graph) PrintLoadOrder(w io.Writer) error {
	order, err := g.LoadOrder()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Порядок загрузки (в пределах выбранного графа):"); err != nil {
		return err
	}
	for i, id := range order {
		if _, err := fmt.Fprintf(w, "%d. %s\n", i+1, id); err != nil {
			return err
		}
	}
	return nil
}
