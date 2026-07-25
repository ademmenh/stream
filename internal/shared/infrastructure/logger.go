package infrastructure

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type levelSplitHandler struct {
	stdout slog.Handler
	stderr slog.Handler
}

func newJSONHandler(w *os.File) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().UTC().Format(time.RFC3339))
			}
			return a
		},
	})
}

func (h *levelSplitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdout.Enabled(ctx, level)
}

func (h *levelSplitHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError {
		return h.stderr.Handle(ctx, r)
	}
	return h.stdout.Handle(ctx, r)
}

func (h *levelSplitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelSplitHandler{stdout: h.stdout.WithAttrs(attrs), stderr: h.stderr.WithAttrs(attrs)}
}

func (h *levelSplitHandler) WithGroup(name string) slog.Handler {
	return &levelSplitHandler{stdout: h.stdout.WithGroup(name), stderr: h.stderr.WithGroup(name)}
}

func InitLogger() {
	slog.SetDefault(slog.New(&levelSplitHandler{
		stdout: newJSONHandler(os.Stdout),
		stderr: newJSONHandler(os.Stderr),
	}))
	slog.Info("app started")
}
