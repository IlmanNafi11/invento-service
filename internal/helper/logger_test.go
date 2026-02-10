package helper

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()

	if logger == nil {
		t.Fatal("NewLogger should not return nil")
	}
	if logger.Logger == nil {
		t.Error("Logger field should not be nil")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{"DEBUG level", DEBUG},
		{"INFO level", INFO},
		{"WARN level", WARN},
		{"ERROR level", ERROR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := &Logger{Logger: NewLogger().Logger}
			logger.Logger.SetOutput(&buf)

			// Test with formatted message
			logger.log(tt.level, "test message %s", "arg1")

			output := buf.String()
			if output == "" {
				t.Errorf("log() should write output")
			}
			if !strings.Contains(output, "test message") {
				t.Errorf("output should contain 'test message', got: %s", output)
			}
			if !strings.Contains(output, string(tt.level)) {
				t.Errorf("output should contain level %s, got: %s", tt.level, output)
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("output should contain DEBUG level")
	}
	if !strings.Contains(output, "debug message") {
		t.Errorf("output should contain 'debug message', got: %s", output)
	}
}

func TestLogger_Debugf(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Debugf("debug %s %d", "message", 123)

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("output should contain DEBUG level")
	}
	if !strings.Contains(output, "debug message 123") {
		t.Errorf("output should contain formatted message, got: %s", output)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Info("info message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("output should contain INFO level")
	}
	if !strings.Contains(output, "info message") {
		t.Errorf("output should contain 'info message', got: %s", output)
	}
}

func TestLogger_Infof(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Infof("info %s %d", "message", 456)

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("output should contain INFO level")
	}
	if !strings.Contains(output, "info message 456") {
		t.Errorf("output should contain formatted message, got: %s", output)
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Warn("warning message")

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Errorf("output should contain WARN level")
	}
	if !strings.Contains(output, "warning message") {
		t.Errorf("output should contain 'warning message', got: %s", output)
	}
}

func TestLogger_Warnf(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Warnf("warning %s %d", "message", 789)

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Errorf("output should contain WARN level")
	}
	if !strings.Contains(output, "warning message 789") {
		t.Errorf("output should contain formatted message, got: %s", output)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Error("error message")

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("output should contain ERROR level")
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("output should contain 'error message', got: %s", output)
	}
}

func TestLogger_Errorf(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Errorf("error %s %d", "message", 999)

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("output should contain ERROR level")
	}
	if !strings.Contains(output, "error message 999") {
		t.Errorf("output should contain formatted message, got: %s", output)
	}
}

func TestLogger_LogFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Info("test message")

	output := buf.String()

	// Check timestamp format [YYYY-MM-DD HH:MM:SS]
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Errorf("output should contain timestamp in brackets, got: %s", output)
	}

	// Check format [timestamp] [level] message
	parts := strings.SplitN(output, "]", 2)
	if len(parts) < 2 {
		t.Errorf("output should have format [timestamp] [level] message, got: %s", output)
	}
}

func TestLogger_WithoutArgs(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	// Test logging without format args
	logger.Debug("simple message")

	output := buf.String()
	if !strings.Contains(output, "simple message") {
		t.Errorf("output should contain 'simple message', got: %s", output)
	}
}

func TestLogger_WithMultipleArgs(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Info("user %s performed action %d with status %s", "john", 42, "success")

	output := buf.String()
	if !strings.Contains(output, "user john performed action 42 with status success") {
		t.Errorf("output should contain formatted message with all args, got: %s", output)
	}
}

func TestLogger_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Info("")

	output := buf.String()
	if output == "" {
		t.Errorf("log should still output timestamp and level even with empty message")
	}
}

func TestLogger_SpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	specialMsg := "Message with special chars: %s, %d, %v"
	logger.Info(specialMsg)

	output := buf.String()
	if !strings.Contains(output, specialMsg) {
		t.Errorf("output should contain special characters, got: %s", output)
	}
}

func TestLogger_UnicodeMessage(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	logger.Info("Pesan bahasa Indonesia: Halo dunia!")

	output := buf.String()
	if !strings.Contains(output, "Pesan bahasa Indonesia") {
		t.Errorf("output should support unicode characters, got: %s", output)
	}
}

func TestLogLevel_Constants(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{"DEBUG constant", DEBUG},
		{"INFO constant", INFO},
		{"WARN constant", WARN},
		{"ERROR constant", ERROR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.level == "" {
				t.Errorf("LogLevel constant should not be empty")
			}
		})
	}
}

func TestLogger_ConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: NewLogger().Logger}
	logger.Logger.SetOutput(&buf)

	// Test concurrent logging
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			logger.Info("concurrent message %d", n)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	output := buf.String()
	// Check that we got some output
	if output == "" {
		t.Errorf("concurrent logging should produce output")
	}
}
