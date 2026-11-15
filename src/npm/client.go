// Package npm получает метаданные пакетов по HTTP без запуска менеджера npm.
package npm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	requestTimeout   = 15 * time.Second
	maxManifestBytes = 16 << 20
)

// Client читает метаданные заданной версии из npm-совместимого реестра.
type Client struct {
	repository string
	httpClient *http.Client
	metadata   map[string]packageMetadata
}

// NewClient создаёт клиент; nil включает стандартный HTTP-клиент с таймаутом.
func NewClient(repository string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: requestTimeout}
	}
	return &Client{repository: strings.TrimRight(repository, "/"), httpClient: httpClient,
		metadata: make(map[string]packageMetadata)}
}

// Dependencies возвращает прямые runtime-зависимости указанной точной версии.
func (c *Client) Dependencies(ctx context.Context, name, version string) ([]Dependency, error) {
	data, err := c.get(ctx, url.PathEscape(name)+"/"+url.PathEscape(version), "application/json")
	if err != nil {
		return nil, err
	}
	return parseManifest(data, name, version)
}

// get читает ограниченный по размеру JSON-ответ реестра.
func (c *Client) get(ctx context.Context, path, accept string) ([]byte, error) {
	endpoint := c.repository + "/" + path
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("URL npm: %w", err)
	}
	request.Header.Set("Accept", accept)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("запрос npm: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("реестр npm вернул HTTP %d для %s", response.StatusCode, path)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxManifestBytes+1))
	if err != nil {
		return nil, fmt.Errorf("чтение ответа npm: %w", err)
	}
	if len(data) > maxManifestBytes {
		return nil, fmt.Errorf("ответ npm превышает лимит 16 МиБ")
	}
	return data, nil
}
