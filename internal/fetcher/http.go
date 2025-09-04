package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPFetcher — реализация Fetcher через net/http с таймаутом.
type HTTPFetcher struct {
	client *http.Client
}

// New создает новый HTTPFetcher с указанным таймаутом.
func New(timeout time.Duration) *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Fetch реализует интерфейс crawler.Fetcher.
func (f *HTTPFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос для %s: %w", url, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko)")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос для %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("неожиданный статус-код %d для %s", resp.StatusCode, url)
	}

	return resp.Body, nil
}
