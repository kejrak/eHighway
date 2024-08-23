package main

import (
	"os"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	TimeFormat string
	Level      zerolog.Level
}

func (l *Logger) loggerInitialize() *zerolog.Logger {
	buildInfo, _ := debug.ReadBuildInfo()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(l.Level).
		With().
		Timestamp().
		Caller().
		Int("pid", os.Getpid()).
		Str("go_version", buildInfo.GoVersion).
		Logger()

	return &logger
}
