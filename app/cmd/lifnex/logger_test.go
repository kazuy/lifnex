package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLoggerDefaultsToTextInfo(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger, err := newLogger(&output, loggerOptions{})
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	logger.Debug("hidden")
	logger.Info("ready", slog.String("transport", "stdio"))

	got := output.String()
	if strings.Contains(got, "hidden") {
		t.Fatalf("output contains debug log: %q", got)
	}
	if !strings.Contains(got, "level=INFO") || !strings.Contains(got, "msg=ready") {
		t.Fatalf("output = %q, want info text log", got)
	}
}

func TestNewLoggerJSONUsesCloudLoggingFields(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger, err := newLogger(&output, loggerOptions{format: "json", level: "error"})
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	logger.Warn("hidden")
	logger.Error("stopped", slog.String("error", "connection closed"))

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if got := entry["severity"]; got != "ERROR" {
		t.Fatalf("severity = %v, want ERROR", got)
	}
	if got := entry["message"]; got != "stopped" {
		t.Fatalf("message = %v, want stopped", got)
	}
	if _, ok := entry["level"]; ok {
		t.Fatalf("entry contains level: %v", entry)
	}
	if _, ok := entry["msg"]; ok {
		t.Fatalf("entry contains msg: %v", entry)
	}
}

func TestNewLoggerRejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options loggerOptions
	}{
		{name: "format", options: loggerOptions{format: "yaml"}},
		{name: "level", options: loggerOptions{level: "verbose"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := newLogger(&bytes.Buffer{}, test.options); err == nil {
				t.Fatal("newLogger() error = nil, want non-nil")
			}
		})
	}
}
