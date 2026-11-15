// Чтение локального тестового репозитория с пакетами A, B, C и т. д.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"mirea/prac2/src/graph"
	"os"
	"regexp"
)

var testName = regexp.MustCompile(`^[A-Z]+$`)

// Test содержит граф, заданный в JSON как имя пакета -> список зависимостей.
type Test struct{ packages map[string][]string }

// LoadTest читает JSON и проверяет все ссылки между пакетами.
func LoadTest(path string) (*Test, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("тестовый репозиторий: %w", err)
	}
	var packages map[string][]string
	if err := json.Unmarshal(data, &packages); err != nil {
		return nil, fmt.Errorf("JSON репозитория: %w", err)
	}
	if packages == nil {
		return nil, fmt.Errorf("репозиторий должен быть JSON-объектом")
	}
	if err := validateTest(packages); err != nil {
		return nil, err
	}
	return &Test{packages: packages}, nil
}

// validateTest проверяет имена пакетов, списки и наличие всех зависимостей.
func validateTest(packages map[string][]string) error {
	for name, deps := range packages {
		if !testName.MatchString(name) || deps == nil {
			return fmt.Errorf("неверная запись тестового пакета %q", name)
		}
		for _, dep := range deps {
			if _, exists := packages[dep]; !exists {
				return fmt.Errorf("пакет %s ссылается на отсутствующий %s", name, dep)
			}
		}
	}
	return nil
}

// Fetch возвращает тестовый пакет; версия в этом формате не используется.
func (r *Test) Fetch(ctx context.Context, name, requirement string) (graph.Node, error) {
	if err := ctx.Err(); err != nil {
		return graph.Node{}, err
	}
	deps, found := r.packages[name]
	if !found {
		return graph.Node{}, fmt.Errorf("тестовый пакет %s не найден", name)
	}
	node := graph.Node{Name: name}
	for _, dep := range deps {
		node.Dependencies = append(node.Dependencies, graph.Dependency{Name: dep})
	}
	return node, nil
}
