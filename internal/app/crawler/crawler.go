package crawler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type Task struct {
	URL       string
	Depth     int
	ParentURL string // URL страницы, на которой была найдена эта ссылка
}

type Crawler struct {
	logger      *slog.Logger
	workerCount int
	maxDepth    int
	sameHost    bool // Ограничить обход только стартовым хостом
	startHost   string
	fetcher     Fetcher
	parser      Parser
	state       State
	storage     Storage
}

func NewCrawler(
	logger *slog.Logger,
	workerCount, maxDepth int,
	sameHost bool,
	fetcher Fetcher,
	storage Storage,
) *Crawler {
	return &Crawler{
		logger:      logger,
		workerCount: workerCount,
		maxDepth:    maxDepth,
		sameHost:    sameHost,
		fetcher:     fetcher,
		parser:      &GoqueryParser{},
		state:       NewSafeMapState(),
		storage:     storage,
	}
}

func (c *Crawler) Run(ctx context.Context, startURL string) error {
	parsedStartURL, err := url.Parse(startURL)
	if err != nil {
		return fmt.Errorf("невалидный стартовый URL %s: %w", startURL, err)
	}
	c.startHost = parsedStartURL.Host

	tasks := make(chan Task, c.workerCount)
	wg := &sync.WaitGroup{}
	g, gCtx := errgroup.WithContext(ctx)

	go func() {
		wg.Wait()
		close(tasks)
	}()

	for i := 0; i < c.workerCount; i++ {
		g.Go(func() error {
			for task := range tasks {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
					c.processAndQueue(gCtx, task, tasks, wg)
				}
			}
			return nil
		})
	}

	// Добавляем первую задачу
	wg.Add(1)
	c.state.MarkVisited(startURL)
	select {
	case tasks <- Task{URL: startURL, Depth: 0, ParentURL: ""}:
	case <-gCtx.Done():
		wg.Done()
	}

	return g.Wait()
}

func (c *Crawler) processAndQueue(ctx context.Context, task Task, tasks chan<- Task, wg *sync.WaitGroup) {
	defer wg.Done()

	log := c.logger.With(slog.String("url", task.URL), slog.Int("depth", task.Depth))

	if task.Depth >= c.maxDepth {
		log.Debug("Достигнута максимальная глубина, задача пропущена")
		return
	}

	log.Info("Обработка страницы")

	body, err := c.fetcher.Fetch(ctx, task.URL)
	if err != nil {
		log.Error("Не удалось загрузить страницу", slog.Any("error", err))
		return
	}
	defer body.Close()

	htmlBytes, err := io.ReadAll(body)
	if err != nil {
		log.Error("Не удалось прочитать тело ответа", slog.Any("error", err))
		return
	}

	links, err := c.parser.ParseLinks(task.URL, string(htmlBytes))
	if err != nil {
		log.Error("Не удалось распарсить страницу", slog.Any("error", err))
		return
	}

	crawledData := CrawledData{
		URL:        task.URL,
		Depth:      task.Depth,
		FoundOn:    task.ParentURL,
		FoundLinks: links,
	}

	saveCtx, saveCancel := context.WithTimeout(ctx, 10*time.Second)
	defer saveCancel()

	if err := c.storage.Save(saveCtx, crawledData); err != nil {
		log.Error("Не удалось сохранить данные", slog.Any("error", err))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, link := range links {
			if c.shouldCrawl(link) && !c.state.IsVisited(link) {
				c.state.MarkVisited(link)
				newTask := Task{URL: link, Depth: task.Depth + 1, ParentURL: task.URL}
				wg.Add(1)
				select {
				case tasks <- newTask:
				case <-ctx.Done():
					wg.Done()
					return
				}
			}
		}
	}()
}

// shouldCrawl проверяет, нужно ли обходить данную ссылку.
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
