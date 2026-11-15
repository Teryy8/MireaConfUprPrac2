// Package config загружает и проверяет семь параметров варианта 17.
package config

import (
	"fmt"
	"io"
)

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

// Print выводит настройки в фиксированном порядке ключ=значение.
func (c Config) Print(w io.Writer) error {
	_, err := fmt.Fprintf(w,
		"package=%s\nrepository=%s\ntest_mode=%t\nversion=%s\noutput=%s\nmax_depth=%d\nfilter=%s\n",
		c.Package, c.Repository, c.TestMode, c.Version, c.Output, c.MaxDepth, c.Filter)
	return err
}
