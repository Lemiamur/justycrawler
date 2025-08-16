package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strimemu/internal/app/crawler"
	"strimemu/internal/config"
)

func main() {
	// Инициализация конфигурации
	cfg, err := config.New()
	if err != nil {
		slog.Error("Ошибка инициализации конфигурации", slog.Any("error", err))
		os.Exit(1)
	}

	// Инициализация логгера
	logLevels := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	level, ok := logLevels[cfg.Log.Level]
	if !ok {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	// Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		logger.Info("Получен сигнал завершения. Завершаю текущие задачи...")
		cancel()
	}()

	// Инициализация зависимостей
	storage, err := crawler.NewMongoStorage(ctx, cfg.Mongo.URI, cfg.Mongo.Database, cfg.Mongo.Collection)
	if err != nil {
		logger.Error("Не удалось подключиться к MongoDB", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := storage.Close(closeCtx); err != nil {
			logger.Error("Не удалось корректно закрыть соединение с MongoDB", slog.Any("error", err))
		}
	}()

	fetcher := crawler.NewHTTPFetcher(cfg.HTTP.Timeout)

	// Инициализация и запуск основной логики
	cr := crawler.NewCrawler(
		logger,
		cfg.WorkerCount,
		cfg.MaxDepth,
		cfg.SameHost,
		fetcher,
		storage,
	)

	logger.Info("Краулер запускается...", slog.Any("config", cfg))

	if err := cr.Run(ctx, cfg.StartURL); err != nil && err != context.Canceled {
		logger.Error("Краулер завершился с ошибкой", slog.Any("error", err))
		os.Exit(1)
	}

	if ctx.Err() == context.Canceled {
		logger.Info("Работа была прервана сигналом завершения.")
	} else {
		logger.Info("Работа успешно завершена.")
	}
}
