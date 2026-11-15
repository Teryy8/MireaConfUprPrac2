// Проверка диапазонов и выбор наибольшей подходящей версии.
package semver

import "fmt"

// Matches проверяет версию с учётом правила npm для prerelease.
func (r Range) Matches(version Version) bool {
	for _, set := range r {
		if matchesSet(version, set) {
			return true
		}
	}
	return false
}

// matchesSet требует выполнения всех сравнений и явного разрешения prerelease.
func matchesSet(version Version, set []comparator) bool {
	allowPre := version.Pre == ""
	for _, c := range set {
		if !compareWith(version, c) {
			return false
		}
		if c.version.Pre != "" && c.version.Numbers == version.Numbers {
			allowPre = true
		}
	}
	return allowPre
}

// compareWith применяет один оператор сравнения.
func compareWith(version Version, c comparator) bool {
	value := Compare(version, c.version)
	switch c.op {
	case "=":
		return value == 0
	case "<":
		return value < 0
	case "<=":
		return value <= 0
	case ">":
		return value > 0
	case ">=":
		return value >= 0
	default:
		return false
	}
}

// Highest выбирает максимальную версию из списка, удовлетворяющую диапазону.
func Highest(versions []string, requirement string) (string, error) {
	rule, err := Compile(requirement)
	if err != nil {
		return "", err
	}
	best := ""
	var bestVersion Version
	for _, text := range versions {
		version, err := Parse(text)
		if err != nil || !rule.Matches(version) {
			continue
		}
		compared := Compare(version, bestVersion)
		if best == "" || compared > 0 || (compared == 0 && text > best) {
			best, bestVersion = text, version
		}
	}
	if best == "" {
		return "", fmt.Errorf("нет версии для условия %q", requirement)
	}
	return best, nil
}
