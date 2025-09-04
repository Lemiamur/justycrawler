package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"justycrawler/internal/app/crawler"
	"justycrawler/internal/config"
	"justycrawler/internal/fetcher"
	"justycrawler/internal/parser"
	"justycrawler/internal/state"
	"justycrawler/internal/storage"
)

const (
	shutdownTimeout = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка выполнения приложения: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Инициализация конфигурации
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("ошибка инициализации конфигурации: %w", err)
	}

	// 2. Инициализация логгера
	logLevels := map[string]slog.Level{
		"debug": slog.LevelDebug, "info": slog.LevelInfo,
		"warn": slog.LevelWarn, "error": slog.LevelError,
	}
	level, ok := logLevels[cfg.Log.Level]
	if !ok {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	// 3. Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		logger.Info("Получен сигнал завершения. Завершаю текущие задачи...")
		cancel()
	}()

	// 4. Инициализация зависимостей
	pageStorage, err := storage.NewMongoStorage(ctx, cfg.Mongo.URI, cfg.Mongo.Database, cfg.Mongo.Collection)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к MongoDB: %w", err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer closeCancel()
		if closeErr := pageStorage.Close(closeCtx); closeErr != nil {
			logger.Error("Не удалось корректно закрыть соединение с MongoDB", slog.Any("error", closeErr))
		}
	}()

	pageState, err := state.NewRedisState(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.SetKey)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}
	defer func() {
		if closeErr := pageState.Close(); closeErr != nil {
			logger.Error("Не удалось корректно закрыть соединение с Redis", slog.Any("error", closeErr))
		}
	}()

	if cfg.ForceRecrawl {
		logger.Info("Флаг --force-recrawl установлен. Очистка состояния в Redis...")
		if err := pageState.Clear(ctx); err != nil {
			return fmt.Errorf("не удалось очистить состояние в Redis: %w", err)
		}
		logger.Info("Состояние успешно очищено.")
	}

	pageFetcher := fetcher.New(cfg.HTTP.Timeout)
	pageParser := parser.New()

	// 5. Инициализация и запуск основной логики
	cr := crawler.NewCrawler(
		logger,
		cfg.WorkerCount,
		cfg.MaxDepth,
		cfg.SameHost,
		pageFetcher,
		pageParser,
		pageStorage,
		pageState,
	)

	logger.Info("Краулер запускается...", slog.Any("config", cfg))

	if err := cr.Run(ctx, cfg.StartURL); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("Краулер завершился с ошибкой", slog.Any("error", err))
		return err
	}

	if errors.Is(ctx.Err(), context.Canceled) {
		logger.Info("Работа была прервана сигналом завершения.")
	} else {
		logger.Info("Работа успешно завершена.")
	}

	return nil
}
