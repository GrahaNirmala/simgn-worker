package logger

import (
	"log/slog"
	"os"
)

type logger struct {
	slog.Logger
}

func newLogger() *logger {
	return &logger{
		Logger: *slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}

var loggerPool = newLogger()

func Log() *logger { return loggerPool }

func (l *logger) Fatal(msg string, args ...any) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}
