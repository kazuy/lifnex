package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type loggerOptions struct {
	format string
	level  string
}

func loadLoggerOptions() loggerOptions {
	return loggerOptions{
		format: os.Getenv("LOG_FORMAT"),
		level:  os.Getenv("LOG_LEVEL"),
	}
}

func newLogger(writer io.Writer, options loggerOptions) (*slog.Logger, error) {
	level, err := parseLogLevel(options.level)
	if err != nil {
		return nil, err
	}

	handlerOptions := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(options.format)) {
	case "", "text":
		handler = slog.NewTextHandler(writer, handlerOptions)
	case "json":
		handlerOptions.ReplaceAttr = replaceCloudLoggingAttr
		handler = slog.NewJSONHandler(writer, handlerOptions)
	default:
		return nil, fmt.Errorf("unsupported log format %q", options.format)
	}

	return slog.New(handler), nil
}

func parseLogLevel(value string) (slog.Level, error) {
	if strings.TrimSpace(value) == "" {
		return slog.LevelInfo, nil
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.TrimSpace(value))); err != nil {
		return 0, fmt.Errorf("failed to parse log level %q: %w", value, err)
	}

	return level, nil
}

func replaceCloudLoggingAttr(_ []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case slog.LevelKey:
		attr.Key = "severity"
	case slog.MessageKey:
		attr.Key = "message"
	}

	return attr
}
