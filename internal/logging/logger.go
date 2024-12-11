// Package logging provides a shared logger and log utilities to be used in all internal packages.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/term"
	"gopkg.in/natefinch/lumberjack.v2"
)

var L = &logger{
	Logger: zerolog.New(zerolog.ConsoleWriter{
		Out:          os.Stderr,
		NoColor:      !isTerminal(),
		PartsExclude: []string{"time"},
		FormatLevel:  consoleFormatLevel,
	}),
}

type logger struct {
	zerolog.Logger
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		short := filepath.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file))
		return fmt.Sprintf("%s:%d", short, line)
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func newLogger(writer io.Writer) *logger {
	return &logger{
		Logger: zerolog.New(writer).With().Timestamp().Caller().Logger(),
	}
}

// UseServerLogger changes L to a logger appropriate for long-running processes,
// like the infra server and connector. If the process is being run in an
// interactive terminal, use the default console logger.
func UseServerLogger() {
	if isTerminal() {
		return
	}
	L = newLogger(os.Stderr)
}

func isTerminal() bool {
	return os.Stdin != nil && term.IsTerminal(int(os.Stdin.Fd()))
}

// UseFileLogger changes L to a logger that writes log output to a file that is
// rotated.
func UseFileLogger(filepath string) {
	zerolog.TimeFieldFormat = time.RFC3339
	writer := &lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    10, // megabytes
		MaxBackups: 7,
		MaxAge:     28, // days
	}

	L = newLogger(writer)
}

type logWriter struct {
	logger interface {
		WithLevel(level zerolog.Level) *zerolog.Event
	}
	level zerolog.Level
}

func (w logWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	p = trimTrailingNewline(p)
	w.logger.WithLevel(w.level).CallerSkipFrame(1).Msg(string(p))
	return n, nil
}

func trimTrailingNewline(p []byte) []byte {
	n := len(p)
	if n > 0 && p[n-1] == '\n' {
		// Trim CR added by stdlog.
		p = p[0 : n-1]
	}
	return p
}

// HTTPErrorLog returns a stdlib log.Logger configured to write logs at
// level to L. The intended use of this logger is for http.Server.ErrorLog.
func HTTPErrorLog(level zerolog.Level) *log.Logger {
	return log.New(logWriter{logger: L, level: level}, "", 0)
}

func Debugf(format string, v ...interface{}) {
	L.Debug().CallerSkipFrame(1).Msgf(format, v...)
}

func Infof(format string, v ...interface{}) {
	L.Info().CallerSkipFrame(1).Msgf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	L.Warn().CallerSkipFrame(1).Msgf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	L.Error().CallerSkipFrame(1).Msgf(format, v...)
}

func SetLevel(levelName string) error {
	level, err := zerolog.ParseLevel(levelName)
	if err != nil {
		return err
	}

	// logging middleware depends on this level. If we stop using the global level
	// make sure to adjust the logging middleware to check the level of the
	// logger instead of the global level.
	zerolog.SetGlobalLevel(level)
	return nil
}

type TestingT interface {
	Cleanup(func())
}

// PatchLogger sets the global L logger to write logs to t. When the test ends
// the global L logger is reset to the previous value.
// PatchLogger changes a static variable, so tests that use PatchLogger can not
// use t.Parallel.
func PatchLogger(t TestingT, writer io.Writer) {
	origL := L
	L = newLogger(writer)
	t.Cleanup(func() {
		L = origL
	})
}
