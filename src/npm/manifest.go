// Разбор JSON-метаданных конкретной версии и объединение прямых зависимостей.
package npm

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Dependency хранит имя зависимости и исходное условие её версии из реестра.
type Dependency struct {
	Name        string
	Requirement string
}

type manifest struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

// parseManifest проверяет ответ и выбирает обычные и необязательные runtime-зависимости.
func parseManifest(data []byte, name, version string) ([]Dependency, error) {
	var metadata manifest
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("JSON метаданных npm: %w", err)
	}
	if metadata.Name != name || metadata.Version != version {
		return nil, fmt.Errorf("метаданные npm не соответствуют %s@%s", name, version)
	}
	dependencies := make(map[string]string)
	for _, group := range []map[string]string{metadata.Dependencies, metadata.OptionalDependencies} {
		for dep, requirement := range group {
			if strings.TrimSpace(dep) == "" || strings.TrimSpace(requirement) == "" {
				return nil, fmt.Errorf("пустое имя или условие версии зависимости в метаданных npm")
			}
			dependencies[dep] = requirement
		}
	}
	names := make([]string, 0, len(dependencies))
	for dep := range dependencies {
		names = append(names, dep)
	}
	sort.Strings(names)
	result := make([]Dependency, 0, len(names))
	for _, dep := range names {
		result = append(result, Dependency{Name: dep, Requirement: dependencies[dep]})
	}
	return result, nil
}
