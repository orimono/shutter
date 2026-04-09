package logger

import (
	"log/slog"
	"os"
)

type LogLevel string

var levelMap = map[LogLevel]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func Init(cfgLevel LogLevel) {
	level, ok := levelMap[cfgLevel]
	if !ok {
		level = slog.LevelInfo
	}

	programLevel := new(slog.LevelVar)
	programLevel.Set(level)

	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
}
