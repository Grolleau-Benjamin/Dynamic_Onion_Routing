package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLogger_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		out         io.Writer
		level       Level
		enableColor bool
		expected    *Logger
	}{
		{
			name:        "standard output, Info level, color enabled",
			out:         os.Stdout,
			level:       Info,
			enableColor: true,
			expected: &Logger{
				out:         os.Stdout,
				enableColor: true,
				timeFormat:  "2006-01-02T15:04:05Z07:00",
				useUTC:      true,
			},
		},
		{
			name:        "nil output, Debug level, color disabled",
			out:         nil,
			level:       Debug,
			enableColor: false,
			expected: &Logger{
				out:         nil,
				enableColor: false,
				timeFormat:  "2006-01-02T15:04:05Z07:00",
				useUTC:      true,
			},
		},
		{
			name:        "file output, Warn level, color enabled",
			out:         os.Stderr,
			level:       Warn,
			enableColor: true,
			expected: &Logger{
				out:         os.Stderr,
				enableColor: true,
				timeFormat:  "2006-01-02T15:04:05Z07:00",
				useUTC:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := New(tt.out, tt.level, tt.enableColor)
			if l.out != tt.expected.out {
				t.Fatalf("New(): Logger.out missmatch \n\tgot: %v\n\twant: %v", l.out, tt.expected.out)
			}

			if l.enableColor != tt.expected.enableColor {
				t.Fatalf("New(): Logger.enableColor missmatch \n\tgot: %v\n\twant: %v", l.enableColor, tt.expected.enableColor)
			}

			if l.timeFormat != tt.expected.timeFormat {
				t.Fatalf("New(): Logger.timeFormat missmatch \n\tgot: %v\n\twant: %v", l.timeFormat, tt.expected.timeFormat)
			}

			if l.useUTC != tt.expected.useUTC {
				t.Fatalf("New(): Logger.useUTC missmatch \n\tgot: %v\n\twant: %v", l.useUTC, tt.expected.useUTC)
			}

			if Level(l.level.Load()) != tt.level {
				t.Fatalf("New(): Logger.level missmatch \n\tgot: %v\n\twant: %v", Level(l.level.Load()), tt.level)
			}

		})
	}
}

func TestLogger_SetOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		newOut   io.Writer
		expected io.Writer
	}{
		{
			name:     "set to stderr",
			newOut:   os.Stderr,
			expected: os.Stderr,
		},
		{
			name:     "set to nil",
			newOut:   nil,
			expected: nil,
		},
		{
			name:     "set to stdout",
			newOut:   os.Stdout,
			expected: os.Stdout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			SetOutput(tt.newOut)

			std.mu.Lock()
			defer std.mu.Unlock()
			if std.out != tt.expected {
				t.Fatalf("SetOutput(): Logger.out missmatch \n\tgot: %v\n\twant: %v", std.out, tt.expected)
			}
		})
	}
}

func TestLogger_SetLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		newLevel Level
		expected Level
	}{
		{
			name:     "set to Debug",
			newLevel: Debug,
			expected: Debug,
		},
		{
			name:     "set to Warn",
			newLevel: Warn,
			expected: Warn,
		},
		{
			name:     "set to Off",
			newLevel: Off,
			expected: Off,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			SetLevel(tt.newLevel)

			level := Level(std.level.Load())
			if level != tt.expected {
				t.Fatalf("SetLevel(): Logger.level missmatch \n\tgot: %v\n\twant: %v", level, tt.expected)
			}
		})
	}
}

func TestLogger_DisableColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "disable color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			DisableColor()

			std.mu.Lock()
			defer std.mu.Unlock()
			if std.enableColor != false {
				t.Fatalf("DisableColor(): Logger.enableColor missmatch \n\tgot: %v\n\twant: %v", std.enableColor, false)
			}
		})
	}
}

func TestLogger_Level_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    Level
		expected string
	}{
		{
			name:     "Debug level",
			level:    Debug,
			expected: "DEBUG",
		},
		{
			name:     "Info level",
			level:    Info,
			expected: "INFO",
		},
		{
			name:     "Warn level",
			level:    Warn,
			expected: "WARN",
		},
		{
			name:     "Error level",
			level:    Error,
			expected: "ERROR",
		},
		{
			name:     "Off level",
			level:    Off,
			expected: "OFF",
		},
		{
			name:     "Unknown level",
			level:    Level(999),
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			str := tt.level.String()
			if str != tt.expected {
				t.Fatalf("Level.String(): missmatch \n\tgot: %v\n\twant: %v", str, tt.expected)
			}
		})
	}
}

func TestLogger_ParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected Level
	}{
		{
			name:     "valid Debug level",
			input:    "DEBUG",
			expected: Debug,
		},
		{
			name:     "valid Info level",
			input:    "INFO",
			expected: Info,
		},
		{
			name:     "valid Warn level",
			input:    "WARN",
			expected: Warn,
		},
		{
			name:     "valid Error level",
			input:    "ERROR",
			expected: Error,
		},
		{
			name:     "valid Off level",
			input:    "OFF",
			expected: Off,
		},
		{
			name:     "invalid level",
			input:    "INVALID",
			expected: Info,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			level := ParseLevel(tt.input)

			if level != tt.expected {
				t.Fatalf("ParseLevel() missmatch \n\tgot: %v\n\twant: %v", level, tt.expected)
			}
		})
	}
}

func TestLogger_output(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		level     Level
		colorCode string
		label     string
		msg       string
		expected  string
	}{
		{
			name:      "Info level message",
			level:     Info,
			colorCode: "32",
			label:     "INFO",
			msg:       "DOR Protocol started",
			expected:  "DOR Protocol started",
		},
		{
			name:      "Debug level message",
			level:     Debug,
			colorCode: "34",
			label:     "DEBUG",
			msg:       "Debugging DOR Protocol",
			expected:  "Debugging DOR Protocol",
		},
		{
			name:      "Warn level message",
			level:     Warn,
			colorCode: "33",
			label:     "WARN",
			msg:       "DOR Protocol warning",
			expected:  "DOR Protocol warning",
		},
		{
			name:      "Error level message",
			level:     Error,
			colorCode: "31",
			label:     "ERROR",
			msg:       "DOR Protocol error occurred",
			expected:  "DOR Protocol error occurred",
		},
		{
			name:      "Fatal level message",
			level:     Error,
			colorCode: "31",
			label:     "FATAL",
			msg:       "DOR Protocol fatal error",
			expected:  "DOR Protocol fatal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			l := New(&buf, tt.level, false)
			l.output(tt.level, tt.colorCode, tt.label, tt.msg)
			output := buf.String()

			if !bytes.Contains([]byte(output), []byte(tt.expected)) {
				t.Fatalf("Logger.output() missmatch \n\tgot: %v\n\twant to contain: %v", output, tt.expected)
			}
		})
	}
}

func TestDebugf_Infof_Warnf_Errorf(t *testing.T) {
	tests := []struct {
		name    string
		logFunc func(string, ...any)
		level   Level
		label   string
		msg     string
	}{
		{
			name:    "Debugf function",
			logFunc: Debugf,
			level:   Debug,
			label:   "DEBUG",
			msg:     "Debugging DOR Protocol",
		},
		{
			name:    "Infof function",
			logFunc: Infof,
			level:   Info,
			label:   "INFO",
			msg:     "DOR Protocol started",
		},
		{
			name:    "Warnf function",
			logFunc: Warnf,
			level:   Warn,
			label:   "WARN",
			msg:     "DOR Protocol warning",
		},
		{
			name:    "Errorf function",
			logFunc: Errorf,
			level:   Error,
			label:   "ERROR",
			msg:     "DOR Protocol error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			oldStd := std
			defer func() { std = oldStd }()

			std = New(&buf, tt.level, false)
			tt.logFunc("%s", tt.msg)

			output := buf.String()

			if !strings.Contains(output, tt.label) {
				t.Errorf("%s: Missing label %q in output: %q", tt.name, tt.label, output)
			}
			if !strings.Contains(output, tt.msg) {
				t.Errorf("%s: Missing message %q in output: %q", tt.name, tt.msg, output)
			}
		})
	}
}
