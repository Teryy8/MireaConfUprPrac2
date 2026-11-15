// Проверки имён пакетов, версий, URL и файловых путей.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

const maxPackageLength = 214

var (
	npmName        = regexp.MustCompile(`^(@[a-z0-9][a-z0-9._-]*/)?[a-z0-9][a-z0-9._-]*$`)
	testName       = regexp.MustCompile(`^[A-Z]+$`)
	versionPattern = regexp.MustCompile(
		`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)` +
			`(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)
	digits = regexp.MustCompile(`^[0-9]+$`)
)

// validate проверяет значения конфигурации.
func (c Config) validate(base string) error {
	if err := c.validatePackage(); err != nil {
		return err
	}
	if !validVersion(c.Version) {
		return fmt.Errorf("version: нужна точная версия, например 4.18.2 или 1.0.0-beta.1")
	}
	if err := c.validateRepository(base); err != nil {
		return err
	}
	if err := validateOutput(resolvePath(base, c.Output)); err != nil {
		return err
	}
	if strings.ContainsFunc(c.Output, unicode.IsControl) {
		return fmt.Errorf("output: управляющие символы недопустимы")
	}
	if c.MaxDepth < 0 {
		return fmt.Errorf("max_depth: глубина должна быть неотрицательной")
	}
	if strings.ContainsFunc(c.Filter, unicode.IsControl) {
		return fmt.Errorf("filter: управляющие символы недопустимы")
	}
	return nil
}

// validatePackage проверяет имя для выбранного режима.
func (c Config) validatePackage() error {
	if c.TestMode {
		if !testName.MatchString(c.Package) {
			return fmt.Errorf("package: в тестовом режиме нужны большие латинские буквы, например A")
		}
	} else if len(c.Package) > maxPackageLength || !npmName.MatchString(c.Package) {
		return fmt.Errorf("package: неверное имя npm-пакета, например express или @scope/name")
	}
	return nil
}

// validVersion проверяет точную SemVer-версию без сторонних библиотек.
func validVersion(version string) bool {
	if !versionPattern.MatchString(version) {
		return false
	}
	core := strings.SplitN(version, "+", 2)[0]
	_, prerelease, found := strings.Cut(core, "-")
	if !found {
		return true
	}
	for _, part := range strings.Split(prerelease, ".") {
		if len(part) > 1 && part[0] == '0' && digits.MatchString(part) {
			return false
		}
	}
	return true
}

// validateRepository проверяет URL реестра или читаемость тестового файла.
func (c Config) validateRepository(base string) error {
	if strings.TrimSpace(c.Repository) == "" || strings.ContainsFunc(c.Repository, unicode.IsControl) {
		return fmt.Errorf("repository: нужен непустой адрес или путь без управляющих символов")
	}
	if c.TestMode {
		return readableFile(resolvePath(base, c.Repository))
	}
	return validateURL(c.Repository)
}

// validateURL проверяет HTTP-адрес реестра без запроса к сети.
func validateURL(address string) error {
	u, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("repository: неверный URL: %w", err)
	}
	if u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return fmt.Errorf("repository: нужен URL с http:// или https://")
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("repository: URL должен быть без логина, query-параметров и фрагмента")
	}
	return nil
}

// readableFile проверяет, что путь ведёт к читаемому обычному файлу.
func readableFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("repository: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("repository: нужен обычный файл")
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("repository: %w", err)
	}
	defer file.Close()
	return nil
}

// validateOutput проверяет расширение SVG и каталог назначения.
func validateOutput(path string) error {
	if strings.ToLower(filepath.Ext(path)) != ".svg" {
		return fmt.Errorf("output: файл должен иметь расширение .svg")
	}
	parent, err := os.Stat(filepath.Dir(path))
	if err != nil || !parent.IsDir() {
		return fmt.Errorf("output: родительский каталог должен существовать")
	}
	info, err := os.Stat(path)
	if err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("output: указан не обычный файл")
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("output: %w", err)
	}
	return nil
}
