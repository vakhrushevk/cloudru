// Package logger предоставляет функцию для инициализации и конфигурирования логгера.
package logger

import (
	"io"
	"log/slog"
	"strings"
)

// ConfigLogger определяет интерфейс конфигурации логгера
// Level() - уровень логирования (debug, info, warn, error)
// Format() - формат вывода (json, text)
// Output() - вывод логов (os.Stdout, os.Stderr, os.File)
type ConfigLogger interface {
	Level() string
	Format() string
	Output() io.Writer
}

// Init инициализирует slog и возвращает логгер, устанавливая его глобально
func Init(config ConfigLogger) *slog.Logger {
	level := parseLevel(config.Level())

	var handler slog.Handler
	switch strings.ToLower(config.Format()) {
	case "text":
		handler = slog.NewTextHandler(config.Output(), &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewJSONHandler(config.Output(), &slog.HandlerOptions{Level: level})
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)
	return logger
}

// parseLevel преобразует строку уровня в slog.Level
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
