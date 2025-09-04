package crawler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"justycrawler/internal/domain"

	"golang.org/x/sync/errgroup"
)

const (
	storageTimeout = 10 * time.Second
)

// Task представляет собой задачу для краулера.
type Task struct {
	URL       string
	Depth     int
	ParentURL string
}

// Crawler представляет собой веб-краулер.
type Crawler struct {
	logger      *slog.Logger
	workerCount int
	maxDepth    int
	sameHost    bool
	startHost   string

	fetcher Fetcher
	parser  Parser
	storage Storage
	state   State
}

// NewCrawler инициализирует новый краулер с внедрением всех зависимостей.
func NewCrawler(
	logger *slog.Logger,
	workerCount, maxDepth int,
	sameHost bool,
	fetcher Fetcher,
	parser Parser,
	storage Storage,
	state State,
) *Crawler {
	return &Crawler{
		logger:      logger,
		workerCount: workerCount,
		maxDepth:    maxDepth,
		sameHost:    sameHost,
		fetcher:     fetcher,
		parser:      parser,
		storage:     storage,
		state:       state,
	}
}

func (c *Crawler) Run(ctx context.Context, startURL string) error {
	parsedStartURL, err := url.Parse(startURL)
	if err != nil {
		return fmt.Errorf("невалидный стартовый URL %s: %w", startURL, err)
	}
	c.startHost = parsedStartURL.Host

	tasks := make(chan Task, c.workerCount)

	g, ctx := errgroup.WithContext(ctx)
	wg := &sync.WaitGroup{}

	for range c.workerCount {
		g.Go(func() error {
			for {
				select {
				case task, ok := <-tasks:
					if !ok {
						return nil
					}
					c.processTask(ctx, task, tasks, wg)
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}

	added, err := c.state.Add(ctx, startURL)
	if err != nil {
		close(tasks)
		return fmt.Errorf("не удалось добавить стартовый URL в стейт: %w", err)
	}

	if added {
		c.logger.InfoContext(ctx, "Добавляем стартовую задачу в очередь.", slog.String("url", startURL))
		wg.Add(1)
		tasks <- Task{URL: startURL, Depth: 0, ParentURL: ""}
	} else {
		c.logger.InfoContext(ctx, "Стартовый URL уже был обработан ранее, новых задач нет.")
	}

	g.Go(func() error {
		wg.Wait()
		close(tasks)
		return nil
	})

	return g.Wait()
}

func (c *Crawler) processTask(ctx context.Context, task Task, tasks chan<- Task, wg *sync.WaitGroup) {
	defer wg.Done()

	log := c.logger.With(slog.String("url", task.URL), slog.Int("depth", task.Depth))
	log.InfoContext(ctx, "Обработка страницы")

	body, err := c.fetcher.Fetch(ctx, task.URL)
	if err != nil {
		log.ErrorContext(ctx, "Не удалось загрузить страницу", slog.Any("error", err))
		return
	}
	defer body.Close()

	htmlBytes, err := io.ReadAll(body)
	if err != nil {
		log.ErrorContext(ctx, "Не удалось прочитать тело ответа", slog.Any("error", err))
		return
	}

	links, err := c.parser.ParseLinks(task.URL, htmlBytes)
	if err != nil {
		log.ErrorContext(ctx, "Не удалось распарсить страницу", slog.Any("error", err))
		return
	}

	c.handleResult(ctx, task, links)

	if task.Depth >= c.maxDepth {
		return
	}

	for _, link := range links {
		if !c.shouldCrawl(link) {
			continue
		}

		added, addErr := c.state.Add(ctx, link)
		if addErr != nil {
			log.ErrorContext(ctx, "Не удалось добавить URL в стейт", slog.Any("error", addErr))
			continue
		}
		if !added {
			log.DebugContext(ctx, "URL уже был обработан ранее.", slog.String("url", link))
			continue
		}

		wg.Add(1)
		select {
		case tasks <- Task{URL: link, Depth: task.Depth + 1, ParentURL: task.URL}:
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

func (c *Crawler) handleResult(ctx context.Context, task Task, foundURLs []string) {
	crawledData := domain.CrawledData{
		URL:        task.URL,
		Depth:      task.Depth,
		FoundOn:    task.ParentURL,
		FoundLinks: foundURLs,
	}

	saveCtx, cancel := context.WithTimeout(ctx, storageTimeout)
	defer cancel()

	if err := c.storage.Save(saveCtx, crawledData); err != nil {
		c.logger.ErrorContext(ctx, "Не удалось сохранить данные",
			slog.String("url", task.URL), slog.Any("error", err))
	}
}

func (c *Crawler) shouldCrawl(link string) bool {
	if !c.sameHost {
		return true
	}
	parsedLink, err := url.Parse(link)
	if err != nil {
		return false
	}
	return parsedLink.Host == c.startHost
}
