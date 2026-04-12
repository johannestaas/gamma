package logx

import (
	"log/slog"
	"os"
)

// Config provides a few values with sane defaults that you can override. The `Level` defaults to `slog.LevelInfo`,
// and the `Out` output file `os.File` will be `os.Stdout`.
type Config struct {
	Level string
	JSON  bool
	Out   *os.File
	// for debugging
	AddSource bool
}

func NewLogger(cfg Config) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	f := cfg.Out
	if f == nil {
		f = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(f, opts)
	} else {
		handler = slog.NewTextHandler(f, opts)
	}

	return slog.New(handler)
}
