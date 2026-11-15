// Package app обрабатывает запуск CLI и вывод параметров первого этапа.
package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"mirea/prac2/src/config"
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
	if err := cfg.Print(out); err != nil {
		fmt.Fprintln(errOut, "Ошибка вывода:", err)
		return exitError
	}
	return exitOK
}
