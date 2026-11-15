package tests

import (
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mirea/prac2/src/render"
)

func TestDOTSharedNodesAndEscaping(t *testing.T) {
	result := buildTestGraph(t, map[string][]string{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": {}}, 3, "")
	dot := render.DOT(result)
	if strings.Count(dot, `"D" [label=`) != 1 {
		t.Fatal("узел D должен быть единственным:", dot)
	}
	for _, edge := range []string{`"A" -> "B"`, `"B" -> "D"`, `"C" -> "D"`} {
		if !strings.Contains(dot, edge) {
			t.Fatal(dot)
		}
	}
	node := result.Nodes["D"]
	node.Name = `quote"\test`
	result.Nodes["D"] = node
	result.Order = []string{`quote"\test`}
	result.Nodes[result.Order[0]] = node
	escaped := render.DOT(result)
	if !strings.Contains(escaped, `"quote\"\\test"`) {
		t.Fatal(escaped)
	}
}

func TestDOTCyclesAndLimits(t *testing.T) {
	cycle := buildTestGraph(t, map[string][]string{"A": {"B"}, "B": {"A"}}, 3, "")
	if !strings.Contains(render.DOT(cycle), `"B" -> "A" [color=red`) {
		t.Fatal(render.DOT(cycle))
	}
	limited := buildTestGraph(t, map[string][]string{"A": {"B"}, "B": {}}, 0, "")
	if !strings.Contains(render.DOT(limited), "предел глубины") {
		t.Fatal(render.DOT(limited))
	}
	empty := buildTestGraph(t, map[string][]string{"A": {}}, 3, "A")
	if !strings.Contains(render.DOT(empty), "исключены фильтром") {
		t.Fatal(render.DOT(empty))
	}
}

func TestSaveSVGAndDOT(t *testing.T) {
	requireGraphviz(t)
	directory := filepath.Join(t.TempDir(), "папка с пробелом")
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "граф.svg")
	result := buildTestGraph(t, map[string][]string{"A": {"B"}, "B": {}}, 3, "")
	if err := render.Save(context.Background(), result, path); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assertSVGGraph(t, raw, 2, 1)
	if _, err := os.Stat(filepath.Join(directory, "граф.dot")); err != nil {
		t.Fatal(err)
	}
}

func TestSVGRendererErrors(t *testing.T) {
	requireGraphviz(t)
	if _, err := render.SVG(context.Background(), "digraph {"); err == nil {
		t.Fatal("нет ошибки Graphviz")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := render.SVG(ctx, "digraph {}"); err == nil {
		t.Fatal("отмена не обработана")
	}
	result := buildTestGraph(t, map[string][]string{"A": {}}, 0, "")
	path := filepath.Join(t.TempDir(), "missing", "graph.svg")
	if err := render.Save(context.Background(), result, path); err == nil {
		t.Fatal("нет ошибки сохранения DOT")
	}
	directory := filepath.Join(t.TempDir(), "graph.svg")
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	if err := render.Save(context.Background(), result, directory); err == nil {
		t.Fatal("нет ошибки сохранения SVG")
	}
	if err := render.Save(ctx, result, path); err == nil {
		t.Fatal("отмена Save не обработана")
	}
}

func TestMissingGraphvizKeepsExistingSVG(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	path := filepath.Join(t.TempDir(), "graph.svg")
	if err := os.WriteFile(path, []byte("old-svg"), 0600); err != nil {
		t.Fatal(err)
	}
	result := buildTestGraph(t, map[string][]string{"A": {}}, 0, "")
	err := render.Save(context.Background(), result, path)
	if err == nil || !strings.Contains(err.Error(), "Graphviz не найден") {
		t.Fatalf("ошибка %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "old-svg" {
		t.Fatalf("старый SVG изменён: %q %v", raw, err)
	}
}

func assertSVGGraph(t *testing.T, raw []byte, wantNodes, wantEdges int) {
	t.Helper()
	var document struct {
		XMLName xml.Name
		Groups  []struct {
			Class string `xml:"class,attr"`
		} `xml:"g>g"`
	}
	if err := xml.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	if document.XMLName.Local != "svg" {
		t.Fatal(document.XMLName)
	}
	nodes, edges := 0, 0
	for _, group := range document.Groups {
		if group.Class == "node" {
			nodes++
		}
		if group.Class == "edge" {
			edges++
		}
	}
	if nodes != wantNodes || edges != wantEdges {
		t.Fatalf("узлов %d, рёбер %d", nodes, edges)
	}
}
