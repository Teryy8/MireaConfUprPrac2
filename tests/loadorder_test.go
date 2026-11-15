package tests

import (
	"bytes"
	"mirea/prac2/src/app"
	"mirea/prac2/src/config"
	"reflect"
	"strings"
	"testing"
)

func TestLoadOrderDependencyFirst(t *testing.T) {
	packages := map[string][]string{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": {}}
	result := buildTestGraph(t, packages, 3, "")
	order, err := result.LoadOrder()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"D", "B", "C", "A"}) {
		t.Fatal(order)
	}
	positions := make(map[string]int)
	for i, id := range order {
		positions[id] = i
	}
	for from, children := range result.Edges {
		for _, to := range children {
			if positions[to] >= positions[from] {
				t.Fatalf("%s раньше %s", from, to)
			}
		}
	}
}

func TestLoadOrderLimitsAndCycles(t *testing.T) {
	packages := map[string][]string{"A": {"B"}, "B": {"C"}, "C": {}}
	for _, tc := range []struct {
		depth  int
		filter string
		want   []string
	}{
		{0, "", []string{"A"}}, {1, "", []string{"B", "A"}}, {5, "B", []string{"A"}}, {5, "A", []string{}},
	} {
		got, err := buildTestGraph(t, packages, tc.depth, tc.filter).LoadOrder()
		if err != nil || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%+v: %v, %v", tc, got, err)
		}
	}
	for _, packages := range []map[string][]string{{"A": {"A"}}, {"A": {"B"}, "B": {"A"}}} {
		if _, err := buildTestGraph(t, packages, 5, "").LoadOrder(); err == nil {
			t.Fatal("не отклонён цикл")
		}
	}
}

func TestLoadOrderOutput(t *testing.T) {
	result := buildTestGraph(t, map[string][]string{"A": {}}, 0, "")
	var out bytes.Buffer
	if err := result.PrintLoadOrder(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1. A") {
		t.Fatal(out.String())
	}
	if err := result.PrintLoadOrder(failedWriter{}); err == nil {
		t.Fatal("нет ошибки записи")
	}
}

func TestLoadOrderConfiguration(t *testing.T) {
	for _, value := range []any{nil, "yes", 17} {
		values := settings()
		values["load_order"] = value
		expectInvalid(t, values, "load_order")
	}
	for _, value := range []bool{false, true} {
		values := settings()
		values["load_order"] = value
		path := writeJSON(t, values)
		cfg, err := config.Load(path)
		if err != nil || cfg.LoadOrder != value {
			t.Fatalf("load_order=%v: %+v, %v", value, cfg, err)
		}
	}
	var out, errOut bytes.Buffer
	code := app.Run([]string{"--config", "../configs/order-diamond.json"}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), "1. D\n2. B\n3. C\n4. A") {
		t.Fatalf("код %d: %s %s", code, &out, &errOut)
	}
}
