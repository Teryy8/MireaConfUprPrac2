package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"mirea/prac2/src/graph"
	"mirea/prac2/src/repository"
)

func testSource(t *testing.T, packages map[string][]string) *repository.Test {
	t.Helper()
	data, err := json.Marshal(packages)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "repository.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	source, err := repository.LoadTest(path)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func buildTestGraph(t *testing.T, packages map[string][]string, depth int, filter string) *graph.Graph {
	t.Helper()
	result, err := graph.Build(context.Background(), testSource(t, packages), "A", "",
		graph.Options{MaxDepth: depth, Filter: filter})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestGraphDFSAndDiamond(t *testing.T) {
	packages := map[string][]string{"A": {"C", "B"}, "B": {"D"}, "C": {"D"}, "D": {}}
	result := buildTestGraph(t, packages, 3, "")
	if !reflect.DeepEqual(result.Order, []string{"A", "B", "D", "C"}) {
		t.Fatal(result.Order)
	}
	if len(result.Cycles) != 0 {
		t.Fatal("общая зависимость принята за цикл")
	}
	if !reflect.DeepEqual(result.Edges["C"], []string{"D"}) {
		t.Fatal(result.Edges)
	}
}

func TestGraphDepthAndFilter(t *testing.T) {
	packages := map[string][]string{"A": {"B", "C"}, "B": {"D"}, "C": {}, "D": {}}
	for depth, count := range map[int]int{0: 1, 1: 3, 2: 4} {
		result := buildTestGraph(t, packages, depth, "")
		if len(result.Nodes) != count {
			t.Fatalf("глубина %d: %+v", depth, result.Nodes)
		}
	}
	result := buildTestGraph(t, packages, 3, "B")
	if len(result.Nodes) != 2 || result.Nodes["B"].Name != "" || result.Nodes["D"].Name != "" {
		t.Fatal(result.Nodes)
	}
	if result := buildTestGraph(t, packages, 3, "A"); len(result.Nodes) != 0 {
		t.Fatal(result.Nodes)
	}
}

func TestGraphCyclesAndSelfLoop(t *testing.T) {
	for _, packages := range []map[string][]string{
		{"A": {"B"}, "B": {"C"}, "C": {"A"}}, {"A": {"A", "A"}},
	} {
		result := buildTestGraph(t, packages, 100, "")
		if len(result.Cycles) != 1 {
			t.Fatal(result.Cycles)
		}
		var output bytes.Buffer
		if err := result.Print(&output); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), "Цикл:") {
			t.Fatal(output.String())
		}
	}
}

func TestGraphShorterPathIsExpanded(t *testing.T) {
	packages := map[string][]string{"A": {"B", "D"}, "B": {"C"}, "C": {"D"}, "D": {"E"}, "E": {}}
	result := buildTestGraph(t, packages, 3, "")
	if result.Depth["D"] != 1 || result.Depth["E"] != 2 {
		t.Fatal(result.Depth)
	}
	if !reflect.DeepEqual(result.Edges["D"], []string{"E"}) {
		t.Fatal(result.Edges)
	}
}

func TestRepositoryInvalidData(t *testing.T) {
	for _, text := range []string{"null", "[]", "{", `{"a":[]}`, `{"A":null}`, `{"A":["B"]}`, `{"A":17}`} {
		path := filepath.Join(t.TempDir(), "repository.json")
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := repository.LoadTest(path); err == nil {
			t.Errorf("принят репозиторий %s", text)
		}
	}
	if _, err := repository.LoadTest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("нет ошибки чтения")
	}
	source := testSource(t, map[string][]string{"B": {}})
	if _, err := graph.Build(context.Background(), source, "A", "", graph.Options{}); err == nil {
		t.Fatal("нет ошибки корня")
	}
}

func TestGraphPrintAndCancellation(t *testing.T) {
	source := testSource(t, map[string][]string{"A": {"B"}, "B": {}})
	result, err := graph.Build(context.Background(), source, "A", "", graph.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := result.Print(failedWriter{}); err == nil {
		t.Fatal("нет ошибки записи")
	}
	var output bytes.Buffer
	if err := result.Print(&output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "предел глубины") {
		t.Fatal(output.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := graph.Build(ctx, source, "A", "", graph.Options{}); err == nil {
		t.Fatal("отмена не обработана")
	}
	if _, err := graph.Build(context.Background(), source, "A", "", graph.Options{MaxDepth: -1}); err == nil {
		t.Fatal("принята отрицательная глубина")
	}
}

func TestGraphCycleAfterShorterPath(t *testing.T) {
	packages := map[string][]string{"A": {"B", "D"}, "B": {"C"}, "C": {"D"}, "D": {"C"}}
	result := buildTestGraph(t, packages, 3, "")
	if !reflect.DeepEqual(result.Edges["C"], []string{"D"}) {
		t.Fatal(result.Edges)
	}
	if !reflect.DeepEqual(result.Edges["D"], []string{"C"}) {
		t.Fatal(result.Edges)
	}
	if len(result.Cycles) != 1 {
		t.Fatalf("цикл C-D не найден: %+v", result.Cycles)
	}
}
