package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"mirea/prac2/src/graph"
	"mirea/prac2/src/npm"
	"mirea/prac2/src/repository"
)

func registryServer(t *testing.T, documents map[string]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := documents[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestNPMGraphVersionsAndCycles(t *testing.T) {
	documents := map[string]string{
		"/express/4.18.2": `{"name":"express","version":"4.18.2","dependencies":{"left":"^1","right":"1.0.0"}}`,
		"/left": `{"name":"left","versions":{"1.0.0":{"name":"left","version":"1.0.0"},` +
			`"1.10.0":{"name":"left","version":"1.10.0","dependencies":{"shared":"1.0.0"}},` +
			`"2.0.0":{"name":"left","version":"2.0.0"}}}`,
		"/right/1.0.0":  `{"name":"right","version":"1.0.0","dependencies":{"shared":"2.0.0"}}`,
		"/shared/1.0.0": `{"name":"shared","version":"1.0.0","dependencies":{"express":"4.18.2"}}`,
		"/shared/2.0.0": `{"name":"shared","version":"2.0.0"}`,
	}
	server := registryServer(t, documents)
	source := repository.NPM{Client: npm.NewClient(server.URL, nil)}
	result, err := graph.Build(context.Background(), source, "express", "4.18.2", graph.Options{MaxDepth: 3})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"express@4.18.2", "left@1.10.0", "shared@1.0.0", "right@1.0.0", "shared@2.0.0"}
	if !reflect.DeepEqual(result.Order, want) {
		t.Fatal(result.Order)
	}
	if len(result.Cycles) != 1 {
		t.Fatal(result.Cycles)
	}
}

func TestNPMGraphFilterBeforeRequest(t *testing.T) {
	server := registryServer(t, map[string]string{
		"/express/4.18.2": `{"name":"express","version":"4.18.2","dependencies":{"skip-this":"^1"}}`,
	})
	source := repository.NPM{Client: npm.NewClient(server.URL, nil)}
	result, err := graph.Build(context.Background(), source, "express", "4.18.2",
		graph.Options{MaxDepth: 5, Filter: "skip"})
	if err != nil || len(result.Nodes) != 1 {
		t.Fatalf("граф %+v, ошибка %v", result, err)
	}
}

func TestNPMResolverTagsAliasesAndCache(t *testing.T) {
	server := registryServer(t, map[string]string{
		"/real": `{"name":"real","dist-tags":{"latest":"1.2.0"},"versions":{` +
			`"1.0.0":{"name":"real","version":"1.0.0"},"1.2.0":{"name":"real","version":"1.2.0"}}}`,
	})
	client := npm.NewClient(server.URL, nil)
	for _, requirement := range []string{"latest", "^1", "npm:real@^1", "npm:real"} {
		name := "real"
		if strings.HasPrefix(requirement, "npm:") {
			name = "alias"
		}
		actual, version, _, err := client.Resolve(context.Background(), name, requirement)
		if err != nil || actual != "real" || version != "1.2.0" {
			t.Fatalf("%s: %s@%s, %v", requirement, actual, version, err)
		}
	}
	server.Close()
	if _, _, _, err := client.Resolve(context.Background(), "real", "~1.2.0"); err != nil {
		t.Fatal("кэш не сработал:", err)
	}
}

func TestNPMResolverErrors(t *testing.T) {
	cases := []struct{ body, requirement string }{
		{`{`, "^1"}, {`null`, "^1"}, {`{"name":"other","versions":{"1.0.0":{}}}`, "^1"},
		{`{"name":"real","versions":{"1.0.0":{"name":"real","version":"1.0.0"}}}`, "^2"},
		{`{"name":"real","versions":{"1.0.0":{}}}`, "^1"},
		{`{"name":"real","dist-tags":{"latest":"2.0.0"},"versions":{"1.0.0":{}}}`, "latest"},
	}
	for _, tc := range cases {
		server := registryServer(t, map[string]string{"/real": tc.body})
		if _, _, _, err := npm.NewClient(server.URL, nil).Resolve(context.Background(), "real", tc.requirement); err == nil {
			t.Errorf("приняты неверные данные %s", tc.body)
		}
	}
	for _, alias := range []string{"npm:", "npm:real@"} {
		client := npm.NewClient("https://example.invalid", nil)
		if _, _, _, err := client.Resolve(context.Background(), "a", alias); err == nil {
			t.Errorf("принят неверный алиас %s", alias)
		}
	}
}
