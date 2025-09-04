package crawler

import (
	"context"
	"io"
	"justycrawler/internal/domain"
)

//go:generate mockery --name Fetcher --output ../../../mocks --outpkg mocks
type Fetcher interface {
	Fetch(ctx context.Context, url string) (io.ReadCloser, error)
}

//go:generate mockery --name Parser --output ../../../mocks --outpkg mocks
type Parser interface {
	ParseLinks(baseRawURL string, htmlBody []byte) ([]string, error)
}

//go:generate mockery --name Storage --output ../../../mocks --outpkg mocks
type Storage interface {
	Save(ctx context.Context, data domain.CrawledData) error
	Close(ctx context.Context) error
}

//go:generate mockery --name State --output ../../../mocks --outpkg mocks
type State interface {
	Add(ctx context.Context, url string) (bool, error)
	Clear(ctx context.Context) error
	Close() error
}
