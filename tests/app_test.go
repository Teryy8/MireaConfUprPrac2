package tests

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mirea/prac2/src/app"
)

func TestCLIOutput(t *testing.T) {
	body := `{"name":"express","version":"4.18.2","dependencies":{"zeta":"~2.0.0","alpha":"^1.0.0"}}`
	path := cliConfig(t, body, http.StatusOK)
	var out, errOut bytes.Buffer
	code := app.Run([]string{"--config", path}, &out, &errOut)
	if code != 0 || errOut.Len() != 0 {
		t.Fatalf("код %d: %s", code, errOut.String())
	}
	expected := "Прямые зависимости express@4.18.2:\n  alpha: ^1.0.0\n  zeta: ~2.0.0\n"
	if out.String() != expected {
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
	if code != 1 || !strings.Contains(errOut.String(), "вывод зависимостей") {
		t.Fatalf("код %d: %s", code, &errOut)
	}
}

func cliConfig(t *testing.T, body string, status int) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
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
	var out, errOut bytes.Buffer
	path := cliConfig(t, `{"name":"express","version":"4.18.2"}`, http.StatusOK)
	if code := app.Run([]string{"--config", path}, &out, &errOut); code != 0 {
		t.Fatalf("код %d: %s", code, &errOut)
	}
	if !strings.Contains(out.String(), "нет зависимостей") {
		t.Fatal(out.String())
	}
}

func TestCLITestModeIsDeferred(t *testing.T) {
	var out, errOut bytes.Buffer
	code := app.Run([]string{"--config", "../configs/test.json"}, &out, &errOut)
	if code != 1 || out.Len() != 0 || !strings.Contains(errOut.String(), "этапе 3") {
		t.Fatalf("код %d: %s", code, &errOut)
	}
}
