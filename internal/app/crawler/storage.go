package crawler

import "context"

type CrawledData struct {
	URL        string   `bson:"url"`
	Depth      int      `bson:"depth"`
	FoundOn    string   `bson:"found_on"`
	FoundLinks []string `bson:"found_links"`
}

type Storage interface {
	Save(ctx context.Context, data CrawledData) error

	// Close закрывает все ресурсы, используемые хранилищем.
	// для вызова этого метода надо использовать свежий контекст
	// (например, с таймаутом), а не тот, который мог быть отменен
	// при завершении основной работы приложения.
	Close(ctx context.Context) error
}
