package log

import (
	"log/slog"
	"os"
	"time"
)

var DefaultLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
	// nolint: revive
	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			a.Value = slog.StringValue(time.Now().Format(time.DateOnly))
		}
		return a
	},
}))

func Error(msg string, args ...any) {
	DefaultLogger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	DefaultLogger.Warn(msg, args...)
}

func Info(msg string, args ...any) {
	DefaultLogger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	DefaultLogger.Debug(msg, args...)
}
