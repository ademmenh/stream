package infrastructure

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func InitLogger(logsDir string) {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		slog.Error("failed to create logs dir", "error", err)
		return
	}

	logFile := filepath.Join(logsDir, "app.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, f)
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().UTC().Format(time.RFC3339))
			}
			return a
		},
	})
	slog.SetDefault(slog.New(handler))
}
