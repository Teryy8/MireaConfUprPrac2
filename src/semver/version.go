// Package semver сравнивает версии и проверяет диапазоны без внешних библиотек.
package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	versionParts     = 3
	maxVersionNumber = 9007199254740991
)

var versionRE = regexp.MustCompile(
	`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)` +
		`(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)

// Version содержит числовую часть версии и prerelease; build не влияет на порядок.
type Version struct {
	Numbers [versionParts]uint64
	Pre     string
}

// Parse проверяет точную SemVer-версию.
func Parse(text string) (Version, error) {
	var result Version
	text = strings.TrimPrefix(text, "v")
	if !versionRE.MatchString(text) {
		return result, fmt.Errorf("неверная версия %q", text)
	}
	core := strings.SplitN(text, "+", 2)[0]
	numbers, pre, _ := strings.Cut(core, "-")
	for i, part := range strings.Split(numbers, ".") {
		number, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return result, fmt.Errorf("неверная версия %q: %w", text, err)
		}
		if number > maxVersionNumber {
			return result, fmt.Errorf("компонент версии слишком велик")
		}
		result.Numbers[i] = number
	}
	for _, part := range strings.Split(pre, ".") {
		if len(part) > 1 && part[0] == '0' && numeric(part) {
			return result, fmt.Errorf("неверный prerelease %q", pre)
		}
	}
	result.Pre = pre
	return result, nil
}

// Compare возвращает -1, 0 или 1 по правилам SemVer.
func Compare(a, b Version) int {
	for i := range a.Numbers {
		if a.Numbers[i] < b.Numbers[i] {
			return -1
		}
		if a.Numbers[i] > b.Numbers[i] {
			return 1
		}
	}
	if a.Pre == b.Pre {
		return 0
	}
	if a.Pre == "" {
		return 1
	}
	if b.Pre == "" {
		return -1
	}
	left, right := strings.Split(a.Pre, "."), strings.Split(b.Pre, ".")
	for i := 0; i < min(len(left), len(right)); i++ {
		if result := compareIdentifier(left[i], right[i]); result != 0 {
			return result
		}
	}
	return compareIdentifier(strconv.Itoa(len(left)), strconv.Itoa(len(right)))
}

// numeric проверяет, состоит ли идентификатор только из цифр.
func numeric(text string) bool {
	if text == "" {
		return false
	}
	for _, digit := range text {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

// compareIdentifier учитывает числовые идентификаторы произвольной длины.
func compareIdentifier(a, b string) int {
	an, bn := numeric(a), numeric(b)
	if an != bn {
		if an {
			return -1
		}
		return 1
	}
	if an && len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}
		return 1
	}
	return strings.Compare(a, b)
}
