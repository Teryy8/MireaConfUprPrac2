// Рекурсивный DFS с защитой от циклов и повторным раскрытием более коротких путей.
package graph

import (
	"context"
	"sort"
)

// visit раскрывает узел, если он впервые найден или достигнут на меньшей глубине.
func (b *builder) visit(ctx context.Context, node Node, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	id := node.ID()
	previous, seen := b.graph.Depth[id]
	if seen && previous <= depth {
		return nil
	}
	if !seen {
		b.graph.Order = append(b.graph.Order, id)
	}
	b.graph.Nodes[id], b.graph.Depth[id] = node, depth
	if depth >= b.options.MaxDepth {
		return nil
	}
	b.active[id] = true
	defer delete(b.active, id)
	deps := append([]Dependency(nil), node.Dependencies...)
	sort.Slice(deps, func(i, j int) bool { return deps[i].Name < deps[j].Name })
	for _, dependency := range deps {
		if err := b.follow(ctx, id, dependency, depth); err != nil {
			return err
		}
	}
	return nil
}

// follow добавляет ребро и продолжает DFS, пока не найден узел текущего стека.
func (b *builder) follow(ctx context.Context, parent string, dependency Dependency, depth int) error {
	if b.excluded(dependency.Name) {
		return nil
	}
	child, err := b.fetch(ctx, dependency.Name, dependency.Requirement)
	if err != nil {
		return err
	}
	if b.excluded(child.Name) {
		return nil
	}
	id := child.ID()
	if !contains(b.graph.Edges[parent], id) {
		b.graph.Edges[parent] = append(b.graph.Edges[parent], id)
	}
	if b.active[id] {
		return nil
	}
	return b.visit(ctx, child, depth+1)
}

// contains проверяет наличие ребра без добавления дубликатов.
func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
