package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type weeklyWriter struct {
	dir      string
	prefix   string
	file     *os.File
	openDate time.Time
	mu       sync.Mutex
}

func (w *weeklyWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil || time.Since(w.openDate) > 7*24*time.Hour {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)
}

func (w *weeklyWriter) rotate() error {
	if w.file != nil {
		w.file.Close()
	}

	name := filepath.Join(w.dir, fmt.Sprintf("%s-%s.log", w.prefix, time.Now().Format("2006-01-02")))
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	w.file = f
	w.openDate = time.Now()
	return nil
}

func (w *weeklyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func Setup(environment, logLevel, logDir string) (*slog.Logger, error) {
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		if environment == "production" {
			level = slog.LevelError
		} else {
			level = slog.LevelInfo
		}
	}

	opts := &slog.HandlerOptions{Level: level}

	if environment == "production" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}

		w := &weeklyWriter{
			dir:    logDir,
			prefix: "app",
		}

		handler := slog.NewJSONHandler(w, opts)
		return slog.New(handler), nil
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler), nil
}
