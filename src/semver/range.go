// Разбор диапазонов npm: сравнения, ^, ~, wildcard, OR и интервал через дефис.
package semver

import (
	"fmt"
	"regexp"
	"strings"
)

type comparator struct {
	op      string
	version Version
}
type partial struct {
	low   Version
	count int
}

// Range хранит объединение наборов сравнений (OR между наборами, AND внутри).
type Range [][]comparator

var operatorSpace = regexp.MustCompile(`([<>=~^]+)\s+`)

// Compile проверяет и разбирает условие версии.
func Compile(text string) (Range, error) {
	var result Range
	for _, branch := range strings.Split(text, "||") {
		set, err := compileBranch(strings.TrimSpace(branch))
		if err != nil {
			return nil, err
		}
		result = append(result, set)
	}
	return result, nil
}

// compileBranch разбирает один набор сравнений или hyphen-интервал.
func compileBranch(text string) ([]comparator, error) {
	if bounds := strings.Split(text, " - "); len(bounds) == 2 {
		return hyphen(bounds[0], bounds[1])
	}
	var result []comparator
	for _, token := range strings.Fields(operatorSpace.ReplaceAllString(text, "$1")) {
		items, err := expand(token)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}

// parsePartial дополняет отсутствующие компоненты нулями и запоминает их число.
func parsePartial(text string) (partial, error) {
	var result partial
	text = strings.TrimPrefix(text, "v")
	core, build, hasBuild := strings.Cut(text, "+")
	core, pre, hasPre := strings.Cut(core, "-")
	normalized, count, err := normalizeNumbers(core)
	if err != nil {
		return result, err
	}
	result.count = count
	if (hasPre || hasBuild) && result.count != versionParts {
		return result, fmt.Errorf("неполная версия %q", text)
	}
	full := strings.Join(normalized, ".")
	if hasPre {
		full += "-" + pre
	}
	if hasBuild {
		full += "+" + build
	}
	version, err := Parse(full)
	result.low = version
	return result, err
}

// upper задаёт исключённую верхнюю границу с обнулёнными младшими компонентами.
func (p partial) upper(index int) Version {
	result := p.low
	result.Numbers[index]++
	for i := index + 1; i < versionParts; i++ {
		result.Numbers[i] = 0
	}
	result.Pre = "0"
	return result
}

// expand преобразует один оператор в обычные сравнения.
func expand(token string) ([]comparator, error) {
	op := "="
	for _, prefix := range []string{"<=", ">=", "<", ">", "=", "^", "~"} {
		if strings.HasPrefix(token, prefix) {
			op, token = prefix, strings.TrimPrefix(token, prefix)
			break
		}
	}
	if token == "" {
		return nil, fmt.Errorf("нет версии после %s", op)
	}
	p, err := parsePartial(token)
	if err != nil {
		return nil, err
	}
	if p.count == 0 {
		return wildcard(op), nil
	}
	if op == "^" || op == "~" {
		return compatible(p, op), nil
	}
	return comparison(p, op), nil
}

// wildcard возвращает условие для полностью произвольной версии.
func wildcard(op string) []comparator {
	if op == "<" || op == ">" {
		return []comparator{{"<", Version{Pre: "0"}}}
	}
	return nil
}

// compatible определяет верхнюю границу для ^ и ~, включая версии 0.x.
func compatible(p partial, op string) []comparator {
	index := 0
	if op == "~" && p.count > 1 {
		index = 1
	}
	if op == "^" {
		for index < p.count-1 && p.low.Numbers[index] == 0 {
			index++
		}
	}
	return []comparator{{">=", p.low}, {"<", p.upper(index)}}
}

// comparison раскрывает сокращённые версии вроде 1.2 и <=1.x.
func comparison(p partial, op string) []comparator {
	if p.count == versionParts {
		return []comparator{{op, p.low}}
	}
	switch op {
	case "=":
		return []comparator{{">=", p.low}, {"<", p.upper(p.count - 1)}}
	case ">":
		upper := p.upper(p.count - 1)
		upper.Pre = ""
		return []comparator{{">=", upper}}
	case "<=":
		return []comparator{{"<", p.upper(p.count - 1)}}
	case "<":
		low := p.low
		low.Pre = "0"
		return []comparator{{"<", low}}
	default:
		return []comparator{{op, p.low}}
	}
}

// hyphen раскрывает включённые границы диапазона A - B.
func hyphen(left, right string) ([]comparator, error) {
	low, err := parsePartial(strings.TrimSpace(left))
	if err != nil {
		return nil, err
	}
	high, err := parsePartial(strings.TrimSpace(right))
	if err != nil {
		return nil, err
	}
	var result []comparator
	if low.count > 0 {
		result = append(result, comparator{">=", low.low})
	}
	if high.count == versionParts {
		result = append(result, comparator{"<=", high.low})
	} else if high.count > 0 {
		result = append(result, comparator{"<", high.upper(high.count - 1)})
	}
	return result, nil
}

// normalizeNumbers заменяет wildcard и отсутствующие компоненты нулями.
func normalizeNumbers(core string) ([]string, int, error) {
	parts := strings.Split(core, ".")
	if len(parts) > versionParts {
		return nil, 0, fmt.Errorf("неверный диапазон %q", core)
	}
	normalized := []string{"0", "0", "0"}
	known, count := true, 0
	for i, part := range parts {
		if isWildcard(part) {
			known = false
			continue
		}
		if !known {
			return nil, 0, fmt.Errorf("неверный wildcard %q", core)
		}
		normalized[i] = part
		count++
	}
	return normalized, count, nil
}

// isWildcard распознаёт обозначения произвольной части версии.
func isWildcard(part string) bool { return part == "*" || strings.EqualFold(part, "x") || part == "" }
