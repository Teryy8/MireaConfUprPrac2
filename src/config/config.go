// Package config загружает и проверяет семь параметров варианта 17.
package config

// Config содержит настройки, заданные в JSON-файле.
type Config struct {
	Package    string `json:"package"`
	Repository string `json:"repository"`
	TestMode   bool   `json:"test_mode"`
	Version    string `json:"version"`
	Output     string `json:"output"`
	MaxDepth   int    `json:"max_depth"`
	Filter     string `json:"filter"`
}
