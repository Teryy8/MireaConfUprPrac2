// Сохранение DOT и запуск внешней программы dot для получения SVG.
package render

import (
	"bytes"
	"context"
	"fmt"
	"mirea/prac2/src/graph"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const renderTimeout = 15 * time.Second

// Save сохраняет DOT рядом с SVG; при ошибке Graphviz новый SVG не записывается.
func Save(ctx context.Context, g *graph.Graph, output string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	source := DOT(g)
	dotPath := strings.TrimSuffix(output, filepath.Ext(output)) + ".dot"
	if err := os.WriteFile(dotPath, []byte(source), 0644); err != nil {
		return fmt.Errorf("сохранение DOT: %w", err)
	}
	svg, err := SVG(ctx, source)
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, svg, 0644); err != nil {
		return fmt.Errorf("сохранение SVG: %w", err)
	}
	return nil
}

// SVG запускает dot -Tsvg без командной оболочки и возвращает изображение.
func SVG(ctx context.Context, source string) ([]byte, error) {
	binary, err := exec.LookPath("dot")
	if err != nil {
		return nil, fmt.Errorf("Graphviz не найден: установите Graphviz и добавьте dot в PATH: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, renderTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-Tsvg")
	command.Stdin = strings.NewReader(source)
	var out, errOut bytes.Buffer
	command.Stdout, command.Stderr = &out, &errOut
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("Graphviz: %w", ctx.Err())
		}
		return nil, fmt.Errorf("Graphviz: %w: %s", err, strings.TrimSpace(errOut.String()))
	}
	return out.Bytes(), nil
}
