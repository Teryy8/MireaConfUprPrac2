package tests

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"mirea/prac2/src/app"
)

func TestCLIOutput(t *testing.T) {
	path := writeJSON(t, settings())
	var out, errOut bytes.Buffer
	code := app.Run([]string{"--config", path}, &out, &errOut)
	if code != 0 || errOut.Len() != 0 {
		t.Fatalf("код %d: %s", code, errOut.String())
	}
	expected := "package=express\nrepository=https://registry.npmjs.org\ntest_mode=false\n" +
		"version=4.18.2\noutput=graph.svg\nmax_depth=3\nfilter=\n"
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
	code := app.Run([]string{"--config", writeJSON(t, settings())}, failedWriter{}, &errOut)
	if code != 1 || !strings.Contains(errOut.String(), "Ошибка вывода") {
		t.Fatalf("код %d: %s", code, &errOut)
	}
}
