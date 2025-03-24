package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	defaultLogger *slog.Logger
	logWriter     *lumberjack.Logger
	once          sync.Once
)

/*
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
*/

type Config struct {
	Level      slog.Level
	Format     string // "json" or "text"
	FilePath   string
	MaxSize    int // megabytes
	MaxBackups int
	MaxAge     int // days
	Compress   bool
	UseConsole bool // also log to console
}

// Usage:
// In main.go:
// err := logger.Initialize(logger.Config{
//     Level:      slog.LevelInfo,
//     Format:     "json",
//     FilePath:   "/var/log/myapp/app.log",
//     MaxSize:    10,
//     MaxBackups: 5,
//     MaxAge:     30,
//     Compress:   true,
//     UseConsole: true, // Also log to console
// })
// In any package:
// func SomeFunction() {
//     logger.Info("Something happened", "key", "value")
//
//     // Or get the logger directly
//     log := logger.Get()
//     log.Debug("Debug info")
// }

func Initialize(config Config) error {
	var err error

	once.Do(func() {
		err = setup(config)
	})

	return err
}

func setup(config Config) error {
	if config.FilePath != "" {
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return err
		}

		logWriter = &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
	}

	var writer io.Writer
	if config.FilePath == "" {
		writer = os.Stdout
	} else if config.UseConsole {
		writer = io.MultiWriter(os.Stdout, logWriter)
	} else {
		writer = logWriter
	}

	// Create the appropriate handler
	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: config.Level,
		})
	} else {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: config.Level,
		})
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	return nil
}

// NOTE: Can use like this: Get().Info("message", "key", "value"); or
// log := logger.Get(); log.Info("message", "key", "value")
func Get() *slog.Logger {
	if defaultLogger == nil {
		return slog.Default()
	}
	return defaultLogger
}

func Close() error {
	if logWriter != nil {
		return logWriter.Close()
	}
	return nil
}

// Helper functions
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// WithValues returns a new logger with the given key-value pairs. Eg:
// userLogger := logger.WithValues("user_id", "12345")
// userLogger.Info("logged in")
// func WithValues(keyValues ...any) *slog.Logger {
// 	logger := Get()
// 	// Convert key-value pairs to attributes
// 	var attrs []slog.Attr
// 	for i := 0; i < len(keyValues); i += 2 {
// 		if i+1 < len(keyValues) {
// 			key, ok := keyValues[i].(string)
// 			if !ok {
// 				continue // Skip if key is not a string
// 			}
// 			attrs = append(attrs, slog.Any(key, keyValues[i+1]))
// 		}
// 	}
// 	return logger.With(attrs...)
// }
