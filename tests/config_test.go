package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mirea/prac2/src/config"
)

func settings() map[string]any {
	return map[string]any{
		"package": "express", "repository": "https://registry.npmjs.org",
		"test_mode": false, "version": "4.18.2", "output": "graph.svg", "max_depth": 3, "filter": "",
	}
}

func writeJSON(t *testing.T, values map[string]any) string {
	t.Helper()
	data, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	return writeConfig(t, data)
}

func writeConfig(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func expectInvalid(t *testing.T, values map[string]any, field string) {
	t.Helper()
	_, err := config.Load(writeJSON(t, values))
	if err == nil || !strings.Contains(err.Error(), field) {
		t.Fatalf("ожидалась ошибка %s, получено %v", field, err)
	}
}

func TestValidConfiguration(t *testing.T) {
	for _, name := range []string{"express", "@scope/package", "a.b-c_d"} {
		values := settings()
		values["package"] = name
		cfg, err := config.Load(writeJSON(t, values))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Package != name || cfg.MaxDepth != 3 {
			t.Fatalf("настройки изменились: %+v", cfg)
		}
	}
}

func TestInvalidParameterValues(t *testing.T) {
	cases := []struct {
		field string
		value any
	}{
		{"package", ""}, {"package", "bad name"}, {"package", "@scope"},
		{"repository", "ftp://example.org"}, {"repository", "https://"},
		{"repository", "https://u:p@example.org"}, {"repository", "https://example.org/?q=x"},
		{"repository", "://bad"}, {"test_mode", "yes"}, {"version", "latest"},
		{"version", "^1.0.0"}, {"version", "01.0.0"}, {"version", "1.0.0-01"},
		{"output", "bad\nname.svg"}, {"output", "graph.png"}, {"output", ""}, {"output", "missing/graph.svg"},
		{"max_depth", -1}, {"max_depth", 1.5}, {"filter", "bad\nfilter"}, {"filter", 15},
	}
	for _, tc := range cases {
		t.Run(tc.field+"/"+strings.ReplaceAll(strings.TrimSpace(toJSON(tc.value)), "/", "_"), func(t *testing.T) {
			values := settings()
			values[tc.field] = tc.value
			expectInvalid(t, values, tc.field)
		})
	}
}

func toJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func TestMissingNullAndWrongType(t *testing.T) {
	for field := range settings() {
		t.Run(field, func(t *testing.T) {
			values := settings()
			delete(values, field)
			expectInvalid(t, values, field)
			values[field] = nil
			expectInvalid(t, values, field)
			values[field] = []string{"wrong"}
			expectInvalid(t, values, field)
		})
	}
}

func TestMalformedJSON(t *testing.T) {
	for _, text := range []string{"", "null", "[]", "{", `{} {}`, `{"package":"x",}`} {
		if _, err := config.Load(writeConfig(t, []byte(text))); err == nil {
			t.Errorf("принят JSON %q", text)
		}
	}
	for _, key := range []string{"unknown", "Package"} {
		values := settings()
		values[key] = "extra"
		expectInvalid(t, values, key)
	}
}

func TestTestRepositoryAndRelativePaths(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "data"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "data", "graph.json")
	if err := os.WriteFile(path, []byte(`{"A":["B"],"B":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	values := settings()
	values["package"], values["test_mode"], values["repository"] = "A", true, "data/graph.json"
	raw, _ := json.Marshal(values)
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(cfgPath); err != nil {
		t.Fatal(err)
	}
	values["repository"] = path
	if _, err := config.Load(writeJSON(t, values)); err != nil {
		t.Fatal(err)
	}
	values["package"] = "express"
	expectInvalid(t, values, "package")
	values["package"], values["repository"] = "A", dir
	expectInvalid(t, values, "repository")
	values["repository"] = "missing.json"
	expectInvalid(t, values, "repository")
}

func TestVersionsAndDepthZero(t *testing.T) {
	for _, version := range []string{"0.0.0", "1.0.0-beta.1", "1.2.3+build.01", "1.0.0-0.a+build"} {
		values := settings()
		values["version"], values["max_depth"] = version, 0
		if _, err := config.Load(writeJSON(t, values)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMissingFileAndOutputDirectory(t *testing.T) {
	if _, err := config.Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("нет ошибки чтения")
	}
	values := settings()
	dir := filepath.Join(t.TempDir(), "folder.svg")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	values["output"] = dir
	expectInvalid(t, values, "output")
}
