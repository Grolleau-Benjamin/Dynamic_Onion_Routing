package logger

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type Level uint8

const (
	Debug Level = iota
	Info
	Warn
	Error
	Off
)

var (
	currentLevel atomic.Uint32
)

func init() {
	currentLevel.Store(uint32(Info))
}

func SetLogLevel(l Level) {
	currentLevel.Store(uint32(l))
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

func allowed(l Level) bool {
	return l >= Level(currentLevel.Load())
}

func log(l Level, color, label, msg string) string {
	if !allowed(l) {
		return ""
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	line := fmt.Sprintf("\033[%sm%s [%s]\033[0m %s", color, ts, label, msg)
	fmt.Println(line)
	return line
}

func LogDebug(format string, a ...any) string {
	return log(Debug, "34", "DEBUG", fmt.Sprintf(format, a...))
}

func LogInfo(format string, a ...any) string {
	return log(Info, "32", "INF", fmt.Sprintf(format, a...))
}

func LogWarning(format string, a ...any) string {
	return log(Warn, "33", "WARN", fmt.Sprintf(format, a...))
}

func LogError(format string, a ...any) string {
	return log(Error, "31", "ERR", fmt.Sprintf(format, a...))
}
