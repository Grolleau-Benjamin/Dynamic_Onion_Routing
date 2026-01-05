package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Level uint32

const (
	Debug Level = iota
	Info
	Warn
	Error
	Off
)

type Logger struct {
	mu          sync.Mutex
	out         io.Writer
	level       atomic.Uint32
	enableColor bool
	timeFormat  string
	useUTC      bool
}

var std = New(os.Stdout, Info, true)

func New(out io.Writer, lvl Level, color bool) *Logger {
	l := &Logger{
		out:         out,
		enableColor: color,
		timeFormat:  time.RFC3339,
		useUTC:      true,
	}
	l.level.Store(uint32(lvl))
	return l
}

func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

func SetLevel(l Level) {
	std.level.Store(uint32(l))
}

func DisableColor() {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.enableColor = false
}

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Off:
		return "OFF"
	default:
		return "UNKNOWN"
	}
}

func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn", "warning":
		return Warn
	case "error":
		return Error
	case "off", "none":
		return Off
	default:
		return Info
	}
}

func (l *Logger) output(lvl Level, colorCode, label, msg string) {
	if lvl < Level(l.level.Load()) {
		return
	}

	now := time.Now()
	if l.useUTC {
		now = now.UTC()
	}
	ts := now.Format(l.timeFormat)

	var line string
	if l.enableColor {
		line = fmt.Sprintf("\033[%sm%s [%-5s]\033[0m %s\n", colorCode, ts, label, msg)
	} else {
		line = fmt.Sprintf("%s [%-5s] %s\n", ts, label, msg)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.out.Write([]byte(line))
}

func Debugf(format string, a ...any) {
	std.output(Debug, "34", "DEBUG", fmt.Sprintf(format, a...))
}

func Infof(format string, a ...any) {
	std.output(Info, "32", "INFO", fmt.Sprintf(format, a...))
}

func Warnf(format string, a ...any) {
	std.output(Warn, "33", "WARN", fmt.Sprintf(format, a...))
}

func Errorf(format string, a ...any) {
	std.output(Error, "31", "ERROR", fmt.Sprintf(format, a...))
}

func Fatalf(format string, a ...any) {
	std.output(Error, "31", "FATAL", fmt.Sprintf(format, a...))
	os.Exit(1)
}
