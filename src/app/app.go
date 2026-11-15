// Package app загружает настройки CLI, выбирает источник и выводит граф зависимостей.
package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"mirea/prac2/src/config"
	"mirea/prac2/src/graph"
	"mirea/prac2/src/npm"
	"mirea/prac2/src/repository"
)

const (
	exitOK    = 0
	exitError = 1
)

// Run читает путь к конфигурации и возвращает код завершения приложения.
func Run(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("npm-graph", flag.ContinueOnError)
	flags.SetOutput(errOut)
	path := flags.String("config", "configs/example.json", "путь к JSON-конфигурации")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitError
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(errOut, "Ошибка: используйте --config путь/к/файлу.json")
		return exitError
	}
	cfg, err := config.Load(*path)
	if err != nil {
		fmt.Fprintln(errOut, "Ошибка:", err)
		return exitError
	}
	if err := runAnalysis(cfg, *path, out); err != nil {
		fmt.Fprintln(errOut, "Ошибка:", err)
		return exitError
	}
	return exitOK
}

// runAnalysis строит и выводит граф из выбранного источника.
func runAnalysis(cfg config.Config, path string, out io.Writer) error {
	source, err := selectSource(cfg, path)
	if err != nil {
		return err
	}
	result, err := graph.Build(context.Background(), source, cfg.Package, cfg.Version,
		graph.Options{MaxDepth: cfg.MaxDepth, Filter: cfg.Filter})
	if err != nil {
		return err
	}
	if err := result.Print(out); err != nil {
		return fmt.Errorf("вывод графа: %w", err)
	}
	if cfg.LoadOrder {
		return result.PrintLoadOrder(out)
	}
	return nil
}

// selectSource открывает тестовый JSON или создаёт клиент реестра npm.
func selectSource(cfg config.Config, path string) (graph.Source, error) {
	if !cfg.TestMode {
		return repository.NPM{Client: npm.NewClient(cfg.Repository, nil)}, nil
	}
	sourcePath := cfg.Repository
	if !filepath.IsAbs(sourcePath) {
		sourcePath = filepath.Join(filepath.Dir(path), sourcePath)
	}
	return repository.LoadTest(sourcePath)
}
