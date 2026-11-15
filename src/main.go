// Программа загружает JSON-настройки анализатора зависимостей npm.
package main

import (
	"mirea/prac2/src/app"
	"os"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
