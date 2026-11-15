package tests

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"mirea/prac2/src/npm"
)

func fetchManifest(t *testing.T, body string) ([]npm.Dependency, error) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/express/4.18.2" {
			t.Errorf("неверный запрос %s", r.URL)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Error("отсутствует Accept")
		}
		_, _ = io.WriteString(w, body)
	}))
	defer server.Close()
	return npm.NewClient(server.URL+"/", nil).Dependencies(context.Background(), "express", "4.18.2")
}

func TestNPMDependencies(t *testing.T) {
	body := `{"name":"express","version":"4.18.2","dependencies":{"zeta":"~2.0.0","alpha":"^1.0.0"},` +
		`"optionalDependencies":{"alpha":"^1.1.0","extra":"*"},` +
		`"devDependencies":{"dev":"1.0.0"},"peerDependencies":{"peer":"2.0.0"}}`
	deps, err := fetchManifest(t, body)
	if err != nil {
		t.Fatal(err)
	}
	want := []npm.Dependency{{Name: "alpha", Requirement: "^1.1.0"}, {Name: "extra", Requirement: "*"},
		{Name: "zeta", Requirement: "~2.0.0"}}
	if !reflect.DeepEqual(deps, want) {
		t.Fatalf("получено %+v, нужно %+v", deps, want)
	}
}

func TestNPMEmptyDependencies(t *testing.T) {
	for _, suffix := range []string{"", `,"dependencies":{}`, `,"dependencies":null`} {
		deps, err := fetchManifest(t, `{"name":"express","version":"4.18.2"`+suffix+`}`)
		if err != nil || len(deps) != 0 {
			t.Fatalf("зависимости %+v, ошибка %v", deps, err)
		}
	}
}

func TestNPMInvalidMetadata(t *testing.T) {
	valid := `{"name":"express","version":"4.18.2"`
	cases := []string{"not JSON", "null", `[]`, `{}`, valid + `}{}`, valid + `,"dependencies":[]}`,
		valid + `,"dependencies":{"x":17}}`, valid + `,"dependencies":{"x":""}}`,
		valid + `,"optionalDependencies":{"":"1.0.0"}}`, `{"name":"other","version":"4.18.2"}`,
		`{"name":"express","version":"5.0.0"}`, `{"name":"express"}`}
	for _, body := range cases {
		if _, err := fetchManifest(t, body); err == nil {
			t.Errorf("приняты некорректные метаданные: %s", body)
		}
	}
}

func TestNPMScopedPackageAndBasePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/registry/@scope%2Fpackage/1.0.0-beta.1+build.2"
		if r.URL.EscapedPath() != want {
			t.Errorf("путь %q вместо %q", r.URL.EscapedPath(), want)
		}
		_, _ = io.WriteString(w, `{"name":"@scope/package","version":"1.0.0-beta.1+build.2"}`)
	}))
	defer server.Close()
	client := npm.NewClient(server.URL+"/registry/", nil)
	_, err := client.Dependencies(context.Background(), "@scope/package", "1.0.0-beta.1+build.2")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNPMResponseLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat(" ", 17<<20))
	}))
	defer server.Close()
	_, err := npm.NewClient(server.URL, nil).Dependencies(context.Background(), "express", "4.18.2")
	if err == nil || !strings.Contains(err.Error(), "лимит") {
		t.Fatalf("ошибка %v", err)
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNPMNetworkFailure(t *testing.T) {
	cause := errors.New("network unavailable")
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) { return nil, cause })}
	_, err := npm.NewClient("https://example.invalid", client).Dependencies(context.Background(), "express", "4.18.2")
	if !errors.Is(err, cause) {
		t.Fatalf("ошибка %v", err)
	}
}

func TestNPMTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	client := npm.NewClient(server.URL, &http.Client{Timeout: 50 * time.Millisecond})
	_, err := client.Dependencies(context.Background(), "express", "4.18.2")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ошибка %v", err)
	}
}

func TestNPMInvalidURL(t *testing.T) {
	_, err := npm.NewClient("://bad", nil).Dependencies(context.Background(), "express", "4.18.2")
	if err == nil {
		t.Fatal("нет ошибки URL")
	}
}

type readFailure struct{}

func (readFailure) Read([]byte) (int, error) { return 0, fmt.Errorf("read failed") }
func (readFailure) Close() error             { return nil }

func TestNPMReadFailure(t *testing.T) {
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: readFailure{}, Header: make(http.Header)}, nil
	})}
	_, err := npm.NewClient("https://example.invalid", client).Dependencies(context.Background(), "express", "4.18.2")
	if err == nil || !strings.Contains(err.Error(), "read failed") {
		t.Fatalf("ошибка %v", err)
	}
}

func TestNPMPrintWriteFailure(t *testing.T) {
	err := npm.Print(failedWriter{}, "express", "4.18.2", nil)
	if err == nil {
		t.Fatal("нет ошибки записи заголовка")
	}
	var out bytes.Buffer
	err = npm.Print(&out, "express", "4.18.2", []npm.Dependency{{Name: "test", Requirement: "1.0.0"}})
	if err != nil || !strings.Contains(out.String(), "test: 1.0.0") {
		t.Fatalf("вывод %q, ошибка %v", out.String(), err)
	}
}
