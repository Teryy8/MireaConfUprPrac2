// Чтение JSON и проверка структуры конфигурации.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Load читает конфигурацию и проверяет типы, обязательные поля и значения.
// Относительные пути внутри JSON отсчитываются от каталога этого JSON-файла.
func Load(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("конфигурация: %w", err)
	}
	if err := checkFields(data); err != nil {
		return cfg, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return cfg, fmt.Errorf("JSON: ожидается один объект")
	}
	if err := cfg.validate(filepath.Dir(path)); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// checkFields проверяет обязательные поля и запрещает null и неизвестные ключи.
func checkFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("JSON: %w", err)
	}
	required := []string{"package", "repository", "test_mode", "version", "output", "max_depth", "filter"}
	for _, key := range required {
		value, exists := fields[key]
		if !exists || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("поле %s обязательно и не может быть null", key)
		}
	}
	for key := range fields {
		found := false
		for _, allowed := range required {
			if key == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("JSON: неизвестное поле %s", key)
		}
	}
	return nil
}

// resolvePath отсчитывает относительный путь от каталога JSON-файла.
func resolvePath(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}
