package tests

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"mirea/prac2/src/app"
)

func TestCLIOutput(t *testing.T) {
	requireGraphviz(t)
	body := `{"name":"express","version":"4.18.2","dependencies":{"zeta":"~2.0.0","alpha":"^1.0.0"}}`
	path := cliConfig(t, body, http.StatusOK)
	var out, errOut bytes.Buffer
	code := app.Run([]string{"--config", path}, &out, &errOut)
	if code != 0 || errOut.Len() != 0 {
		t.Fatalf("код %d: %s", code, errOut.String())
	}
	expected := "Граф зависимостей:\nexpress@4.18.2 -> alpha@1.0.0, zeta@2.0.0\n" +
		"alpha@1.0.0 -> (нет зависимостей)\nzeta@2.0.0 -> (нет зависимостей)\n"
	if !strings.HasPrefix(out.String(), expected) || !strings.Contains(out.String(), "SVG:") {
		t.Fatalf("неверный вывод: %q", out.String())
	}
}

func TestCLIErrorsAndHelp(t *testing.T) {
	cases := []struct {
		args    []string
		code    int
		message string
	}{
		{[]string{"--config", "not-found.json"}, 1, "Ошибка"},
		{[]string{"--unknown"}, 1, "flag provided"},
		{[]string{"--config"}, 1, "needs an argument"},
		{[]string{"extra"}, 1, "Ошибка"},
		{[]string{"--help"}, 0, "config"},
	}
	for _, tc := range cases {
		var out, errOut bytes.Buffer
		code := app.Run(tc.args, &out, &errOut)
		if code != tc.code || !strings.Contains(errOut.String(), tc.message) || out.Len() != 0 {
			t.Fatalf("аргументы %v: код %d, stdout %q, stderr %q", tc.args, code, out.String(), errOut.String())
		}
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestCLIWriteError(t *testing.T) {
	var errOut bytes.Buffer
	path := cliConfig(t, `{"name":"express","version":"4.18.2"}`, http.StatusOK)
	code := app.Run([]string{"--config", path}, failedWriter{}, &errOut)
	if code != 1 || !strings.Contains(errOut.String(), "вывод графа") {
		t.Fatalf("код %d: %s", code, &errOut)
	}
}

func cliConfig(t *testing.T, body string, status int) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		payload := body
		for _, name := range []string{"alpha", "zeta"} {
			if r.URL.Path == "/"+name {
				version := "1.0.0"
				if name == "zeta" {
					version = "2.0.0"
				}
				payload = fmt.Sprintf(`{"name":%q,"versions":{%q:{"name":%q,"version":%q}}}`, name, version, name, version)
			}
		}
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(server.Close)
	values := settings()
	values["repository"] = server.URL
	return writeJSON(t, values)
}

func TestCLIRegistryErrors(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusTooManyRequests, http.StatusInternalServerError} {
		var out, errOut bytes.Buffer
		path := cliConfig(t, `{}`, status)
		code := app.Run([]string{"--config", path}, &out, &errOut)
		if code != 1 || out.Len() != 0 || !strings.Contains(errOut.String(), "HTTP") {
			t.Fatalf("HTTP %d: код %d, stdout %q, stderr %q", status, code, out.String(), errOut.String())
		}
	}
}

func TestCLIEmptyDependencies(t *testing.T) {
	requireGraphviz(t)
	var out, errOut bytes.Buffer
	path := cliConfig(t, `{"name":"express","version":"4.18.2"}`, http.StatusOK)
	if code := app.Run([]string{"--config", path}, &out, &errOut); code != 0 {
		t.Fatalf("код %d: %s", code, &errOut)
	}
	if !strings.Contains(out.String(), "нет зависимостей") {
		t.Fatal(out.String())
	}
}

func TestCLITestMode(t *testing.T) {
	requireGraphviz(t)
	var out, errOut bytes.Buffer
	values := settings()
	absolute, err := filepath.Abs("../examples/repository.json")
	if err != nil {
		t.Fatal(err)
	}
	values["package"], values["test_mode"], values["repository"] = "A", true, absolute
	path := writeJSON(t, values)
	code := app.Run([]string{"--config", path}, &out, &errOut)
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "graph.svg")); err != nil {
		t.Fatal(err)
	}
	if code != 0 || errOut.Len() != 0 || !strings.Contains(out.String(), "A -> B, C") {
		t.Fatalf("код %d: %s", code, &errOut)
	}
}

func requireGraphviz(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("dot"); err != nil {
		t.Skip("проверка SVG требует Graphviz")
	}
}
