// Package app загружает настройки CLI и выводит прямые зависимости npm-пакета.
package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"mirea/prac2/src/config"
	"mirea/prac2/src/npm"
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
	if err := runAnalysis(cfg, out); err != nil {
		fmt.Fprintln(errOut, "Ошибка:", err)
		return exitError
	}
	return exitOK
}

// runAnalysis получает и выводит прямые зависимости без обхода графа.
func runAnalysis(cfg config.Config, out io.Writer) error {
	if cfg.TestMode {
		return fmt.Errorf("тестовый репозиторий будет поддержан на этапе 3; укажите test_mode=false")
	}
	client := npm.NewClient(cfg.Repository, nil)
	dependencies, err := client.Dependencies(context.Background(), cfg.Package, cfg.Version)
	if err != nil {
		return err
	}
	if err := npm.Print(out, cfg.Package, cfg.Version, dependencies); err != nil {
		return fmt.Errorf("вывод зависимостей: %w", err)
	}
	return nil
}
