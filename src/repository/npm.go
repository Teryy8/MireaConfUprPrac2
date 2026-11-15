// Package repository предоставляет источники пакетов для обхода графа.
package repository

import (
	"context"
	"mirea/prac2/src/graph"
	"mirea/prac2/src/npm"
)

// NPM преобразует метаданные реестра в узлы графа.
type NPM struct{ Client *npm.Client }

// Fetch подбирает версию зависимости и получает сведения о ней.
func (r NPM) Fetch(ctx context.Context, name, requirement string) (graph.Node, error) {
	actualName, version, deps, err := r.Client.Resolve(ctx, name, requirement)
	if err != nil {
		return graph.Node{}, err
	}
	node := graph.Node{Name: actualName, Version: version}
	for _, dep := range deps {
		node.Dependencies = append(node.Dependencies, graph.Dependency{Name: dep.Name, Requirement: dep.Requirement})
	}
	return node, nil
}
