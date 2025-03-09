package service

import (
	"log/slog"
	"os"
	"path/filepath"
)

// InitLogging initializes an embedded slog Logger.
func InitLogging(logLevelStr, logFormat string, toFile bool) error {

	level := slog.LevelInfo
	if err := level.UnmarshalText([]byte(logLevelStr)); err != nil {
		return err
	}

	logFile := os.Stderr
	if toFile {
		// TODO
	}

	handlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			// Shorten file path for source logging
			if attr.Key == slog.SourceKey {
				if source, ok := attr.Value.Any().(*slog.Source); ok {
					source.File = filepath.Base(source.File)
				}
			}
			return attr
		},
	}

	// Base TextHandler with file output
	var baseHandler slog.Handler
	switch logFormat {
	case "json":
		baseHandler = slog.NewJSONHandler(logFile, handlerOpts)
	default:
		baseHandler = slog.NewTextHandler(logFile, handlerOpts)
	}

	// Set as default logger
	slog.SetDefault(slog.New(baseHandler))

	return nil
}
