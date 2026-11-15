// Выбор конкретной версии по метаданным npm и условиям из dependencies.
package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"mirea/prac2/src/semver"
)

type packageMetadata struct {
	Name     string                     `json:"name"`
	Versions map[string]json.RawMessage `json:"versions"`
	Tags     map[string]string          `json:"dist-tags"`
}

// Resolve получает точную версию и её зависимости; поддерживает диапазоны, теги и npm-алиасы.
func (c *Client) Resolve(ctx context.Context, name, requirement string) (string, string, []Dependency, error) {
	name, requirement, err := alias(name, requirement)
	if err != nil {
		return "", "", nil, err
	}
	if _, err := semver.Parse(requirement); err == nil {
		requirement = strings.TrimPrefix(requirement, "v")
		deps, err := c.Dependencies(ctx, name, requirement)
		return name, requirement, deps, err
	}
	metadata, err := c.packageVersions(ctx, name)
	if err != nil {
		return "", "", nil, err
	}
	version, err := selectVersion(metadata, requirement)
	if err != nil {
		return "", "", nil, fmt.Errorf("%s: %w", name, err)
	}
	deps, err := parseManifest(metadata.Versions[version], name, version)
	return name, version, deps, err
}

// alias отделяет настоящее имя пакета в npm:package@range.
func alias(name, requirement string) (string, string, error) {
	if !strings.HasPrefix(requirement, "npm:") {
		return name, requirement, nil
	}
	target := strings.TrimPrefix(requirement, "npm:")
	index := strings.LastIndex(target, "@")
	if target == "" || index == len(target)-1 {
		return "", "", fmt.Errorf("неверный npm-алиас %q", requirement)
	}
	if index <= 0 {
		return target, "latest", nil
	}
	return target[:index], target[index+1:], nil
}

// packageVersions кэширует сокращённые метаданные пакета на время одного анализа.
func (c *Client) packageVersions(ctx context.Context, name string) (packageMetadata, error) {
	if cached, found := c.metadata[name]; found {
		return cached, nil
	}
	var metadata packageMetadata
	data, err := c.get(ctx, url.PathEscape(name), "application/vnd.npm.install-v1+json")
	if err != nil {
		return metadata, err
	}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return metadata, fmt.Errorf("JSON версий npm: %w", err)
	}
	if metadata.Name != name || len(metadata.Versions) == 0 {
		return metadata, fmt.Errorf("нет версий пакета %s", name)
	}
	c.metadata[name] = metadata
	return metadata, nil
}

// selectVersion сначала ищет тег, затем максимальную версию по диапазону.
func selectVersion(metadata packageMetadata, requirement string) (string, error) {
	if version, found := metadata.Tags[requirement]; found {
		if _, exists := metadata.Versions[version]; !exists {
			return "", fmt.Errorf("тег указывает на отсутствующую версию")
		}
		return version, nil
	}
	versions := make([]string, 0, len(metadata.Versions))
	for version := range metadata.Versions {
		versions = append(versions, version)
	}
	return semver.Highest(versions, requirement)
}
