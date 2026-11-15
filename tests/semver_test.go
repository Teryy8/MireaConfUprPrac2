package tests

import (
	"mirea/prac2/src/semver"
	"strings"
	"testing"
)

func TestSemverRanges(t *testing.T) {
	cases := []struct {
		rule, version string
		want          bool
	}{
		{"^1.2.3", "1.9.0", true}, {"^1.2.3", "2.0.0", false},
		{"^0.2.3", "0.3.0", false}, {"^0.0.3", "0.0.4", false},
		{"^0.0", "0.0.9", true}, {"^0", "0.9.0", true},
		{"~1.2.3", "1.3.0", false}, {"~1", "1.9.0", true},
		{"1.x", "1.9.9", true}, {"1.2", "1.2.9", true}, {"1.2", "1.3.0", false},
		{">= 1.0.0 < 2.0.0", "1.5.0", true}, {">=1 <2", "2.0.0", false},
		{"1 || 3", "2.0.0", false}, {"1 || 3", "3.2.1", true},
		{"1.2.3 - 2.3", "2.3.99", true}, {"1.2 - 2.3.4", "2.3.4", true},
		{"<=1.2", "1.2.99", true}, {">1.2", "1.3.0", true},
		{"<1.2", "1.2.0", false}, {"<=1.2.3", "1.2.3", true},
		{"*", "1.0.0-beta", false}, {"", "1.0.0", true}, {">*", "1.0.0", false},
		{">=1.0.0-beta.1 <2.0.0", "1.0.0-beta.2", true},
		{">=1.0.0-beta.1 <2.0.0", "1.1.0-beta.1", false},
		{"v1.2.3", "1.2.3+build.01", true}, {"1.0.0-beta", "1.0.0-beta", true},
	}
	for _, tc := range cases {
		t.Run(tc.rule+"/"+tc.version, func(t *testing.T) {
			rule, err := semver.Compile(tc.rule)
			if err != nil {
				t.Fatal(err)
			}
			version, err := semver.Parse(tc.version)
			if err != nil {
				t.Fatal(err)
			}
			if got := rule.Matches(version); got != tc.want {
				t.Fatalf("получено %v, нужно %v", got, tc.want)
			}
		})
	}
}

func TestSemverOrderAndSelection(t *testing.T) {
	ordered := []string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2",
		"1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0"}
	for i := 1; i < len(ordered); i++ {
		a, _ := semver.Parse(ordered[i-1])
		b, _ := semver.Parse(ordered[i])
		if semver.Compare(a, b) >= 0 || semver.Compare(b, a) <= 0 {
			t.Fatalf("порядок %s, %s", ordered[i-1], ordered[i])
		}
	}
	versions := []string{"1.2.0", "1.10.0", "1.9.0", "2.0.0", "broken", "1.11.0-beta"}
	best, err := semver.Highest(versions, "^1.0.0")
	if err != nil || best != "1.10.0" {
		t.Fatalf("версия %s, ошибка %v", best, err)
	}
	if _, err := semver.Highest(versions, "^3"); err == nil {
		t.Fatal("нет ошибки отсутствия версии")
	}
}

func TestSemverInvalidInputs(t *testing.T) {
	for _, text := range []string{"01.0.0", "1.0.0-01", "1.0.0+", "1.0", "1.0.0-", "9007199254740992.0.0"} {
		if _, err := semver.Parse(text); err == nil {
			t.Errorf("принята версия %q", text)
		}
	}
	for _, text := range []string{"^", "hello", "file:../x", "git+https://example.com/a", "1.2.3.4", "1.x.3"} {
		if _, err := semver.Compile(text); err == nil {
			t.Errorf("принят диапазон %q", text)
		}
	}
	a, _ := semver.Parse("1.0.0-" + strings.Repeat("9", 100))
	b, _ := semver.Parse("1.0.0-" + strings.Repeat("9", 99))
	if semver.Compare(a, b) <= 0 {
		t.Fatal("неверное сравнение длинных чисел prerelease")
	}
}
