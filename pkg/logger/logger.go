package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger provides leveled logging.
type Logger struct {
	*zap.SugaredLogger
}

// Options holds the logger settings.
type Options struct {
	Level string // minimum log level (debug/info/warn/error)
	File  string // log file path, empty disables file output
}

func New(opts Options) (*Logger, error) {
	level, err := opts.parseLogLevel()
	if err != nil {
		return nil, err
	}

	cores := []zapcore.Core{
		newStreamCore(os.Stdout, level, zapcore.WarnLevel),
		newStreamCore(os.Stderr, max(level, zapcore.ErrorLevel), zapcore.FatalLevel),
	}

	if opts.File != "" {
		core, err := newFileCore(opts.File, level)
		if err != nil {
			return nil, err
		}
		cores = append(cores, core)
	}

	base := zap.New(zapcore.NewTee(cores...), zap.AddCaller())
	return &Logger{SugaredLogger: base.Sugar()}, nil
}

// newStreamCore builds a console core.
func newStreamCore(w io.Writer, minLevel, maxLevel zapcore.Level) zapcore.Core {
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	enabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= minLevel && lvl <= maxLevel
	})
	return zapcore.NewCore(encoder, zapcore.AddSync(w), enabler)
}

// newFileCore builds a file core.
func newFileCore(filePath string, level zapcore.Level) (zapcore.Core, error) {
	if dir := filepath.Dir(filePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("[logger] create log dir %s error: %w", dir, err)
		}
	}

	writer := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    100,  // megabytes per file (default to 100)
		MaxBackups: 7,    // rotated files to keep (default to 7)
		MaxAge:     30,   // days to keep a rotated file (default to 30)
		Compress:   true, // gzip rotated files to save disk
	}

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	return zapcore.NewCore(encoder, zapcore.AddSync(writer), level), nil
}

// parseLogLevel parses the logger level from the configured level string.
func (o Options) parseLogLevel() (zapcore.Level, error) {
	switch strings.ToLower(o.Level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "", "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return 0, fmt.Errorf("[logger] invalid level %q: must be debug, info, warn or error", o.Level)
	}
}
