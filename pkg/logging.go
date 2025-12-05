package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type MultiIoWriter struct {
	writers []io.Writer
}

func (m *MultiIoWriter) Write(p []byte) (n int, err error) {
	var done chan bool
	for _, w := range m.writers {
		go func() {
			_, err := w.Write(p)
			done <- true
			if err != nil {
				return
			}
		}()
	}
	for range m.writers {
		<-done
	}
	return len(p), nil
}

func InitLogging() {
	if Config.LogFormat == "json" {
		// fluent-bit on Windows doesn't support stdin, so we dual write to a file
		f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(fmt.Errorf("failed to open log file: %w", err))
		}
		defer f.Close()
		handler := slog.NewJSONHandler(io.MultiWriter(os.Stdout, f), &slog.HandlerOptions{Level: slog.Level(-4)})
		base := slog.New(handler)
		defaultLogger := base.With(slog.String("env", Config.AppEnv))
		slog.SetDefault(defaultLogger)
	}
}
