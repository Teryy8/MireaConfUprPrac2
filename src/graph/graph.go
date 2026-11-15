// Package graph строит граф зависимостей рекурсивным обходом в глубину.
package graph

import (
	"context"
	"fmt"
	"strings"
)

// Dependency задаёт имя пакета и условие его версии.
type Dependency struct{ Name, Requirement string }

// Node представляет конкретную версию пакета с её прямыми зависимостями.
type Node struct {
	Name, Version string
	Dependencies  []Dependency
}

// ID различает версии одного пакета; в тестовом репозитории версия отсутствует.
func (n Node) ID() string {
	if n.Version == "" {
		return n.Name
	}
	return n.Name + "@" + n.Version
}

// Source получает пакет из npm или локального тестового файла.
type Source interface {
	Fetch(context.Context, string, string) (Node, error)
}

// Options задаёт максимальную глубину и фильтр по имени.
type Options struct {
	MaxDepth int
	Filter   string
}

// Edge хранит ребро графа, в том числе обратное ребро цикла.
type Edge struct{ From, To string }

// Graph хранит узлы, рёбра, порядок DFS и обнаруженные обратные рёбра.
type Graph struct {
	Root   string
	Nodes  map[string]Node
	Edges  map[string][]string
	Order  []string
	Cycles []Edge
	Depth  map[string]int
	Limit  int
}

type builder struct {
	source  Source
	options Options
	graph   *Graph
	active  map[string]bool
	cache   map[string]Node
}

// Build строит граф, исключая отфильтрованные пакеты до запроса к источнику.
func Build(ctx context.Context, source Source, name, version string, options Options) (*Graph, error) {
	if options.MaxDepth < 0 {
		return nil, fmt.Errorf("глубина не может быть отрицательной")
	}
	result := &Graph{Nodes: make(map[string]Node), Edges: make(map[string][]string),
		Depth: make(map[string]int), Limit: options.MaxDepth}
	b := &builder{source: source, options: options, graph: result,
		active: make(map[string]bool), cache: make(map[string]Node)}
	if b.excluded(name) {
		return result, nil
	}
	root, err := b.fetch(ctx, name, version)
	if err != nil {
		return nil, err
	}
	result.Root = root.ID()
	if err := b.visit(ctx, root, 0); err != nil {
		return nil, err
	}
	result.detectCycles()
	return result, nil
}

// excluded проверяет фильтр; пустая строка не исключает пакеты.
func (b *builder) excluded(name string) bool {
	return b.options.Filter != "" && strings.Contains(name, b.options.Filter)
}

// fetch кэширует уже полученные узлы и условия версий.
func (b *builder) fetch(ctx context.Context, name, version string) (Node, error) {
	key := name + "\x00" + version
	if cached, ok := b.cache[key]; ok {
		return cached, nil
	}
	node, err := b.source.Fetch(ctx, name, version)
	if err != nil {
		return node, fmt.Errorf("%s: %w", name, err)
	}
	b.cache[key] = node
	b.cache[node.Name+"\x00"+node.Version] = node
	return node, nil
}
