// Вывод прямых зависимостей в консоль в алфавитном порядке.
package npm

import (
	"fmt"
	"io"
)

// Print выводит прямые зависимости или сообщение об их отсутствии.
func Print(w io.Writer, name, version string, dependencies []Dependency) error {
	if _, err := fmt.Fprintf(w, "Прямые зависимости %s@%s:\n", name, version); err != nil {
		return err
	}
	if len(dependencies) == 0 {
		_, err := fmt.Fprintln(w, "  нет зависимостей")
		return err
	}
	for _, dep := range dependencies {
		if _, err := fmt.Fprintf(w, "  %s: %s\n", dep.Name, dep.Requirement); err != nil {
			return err
		}
	}
	return nil
}
