package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	zl *zerolog.Logger
}

var logger Logger

func init() {
	zl := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Caller().
		Timestamp().
		Logger()

	logger = Logger{
		zl: &zl,
	}
}

func New() Logger {
	return logger
}

func GlobalSetLogFile(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	newZl := logger.zl.Output(f)

	*logger.zl = newZl

	return nil
}

func (l Logger) Info() *zerolog.Event {
	return l.zl.Info()
}

func (l Logger) Debug() *zerolog.Event {
	return l.zl.Debug()
}

func (l Logger) Warn() *zerolog.Event {
	return l.zl.Warn()
}

func (l Logger) Error() *zerolog.Event {
	return l.zl.Error()
}

func (l Logger) Fatal() *zerolog.Event {
	return l.zl.Fatal()
}
