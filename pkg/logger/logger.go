package logger

import (
	"errors"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"runtime"
	"sync/atomic"
)

const (
	JsonFormat = "json"
	TextFormat = "text"
)

const (
	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelPanic = "panic"
	LevelFatal = "fatal"
)

var defaultLogger *Logger

func init() {
	Setup(defaultConfig)
}

func Setup(config Config) error {
	logger, err := newLogger(config)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

func newLogger(config Config) (*Logger, error) {

	if len(config.Level) == 0 {
		config.Level = defaultConfig.Level
	}

	if len(config.Format) == 0 {
		config.Format = defaultConfig.Format
	}

	var (
		logger       = new(Logger)
		hooks        []HookCloser
		logrusLogger = logrus.New()
	)

	if len(config.InfoLog) > 0 {
		// setup hooks
		infoHook, err := newLevelFileHook(config.InfoLog, logrus.InfoLevel, logrus.WarnLevel)
		if err != nil {
			return logger, err
		}
		hooks = append(hooks, infoHook)
	}

	if len(config.ErrorLog) > 0 {
		errorHook, err := newLevelFileHook(config.ErrorLog, logrus.ErrorLevel)
		if err != nil {
			return logger, err
		}
		hooks = append(hooks, errorHook)
	}

	for _, hook := range hooks {
		logrusLogger.AddHook(hook)
	}

	// setup formatter
	if config.Format == JsonFormat {
		logrusLogger.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				return "", ""
			},
		}
	} else {
		logrusLogger.Formatter = &nested.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			HideKeys:        true,
			NoColors:        true,
			NoFieldsColors:  true,
			TrimMessages:    true,
			CallerFirst:     false,
			CustomCallerFormatter: func(frame *runtime.Frame) string {
				return ""
			},
		}
	}

	logger.hs = hooks
	logger.logger = logrusLogger

	return logger, nil
}

// default logger configuration
var defaultConfig = Config{
	Level:    "info",
	Format:   "text",
	InfoLog:  "",
	ErrorLog: "",
}

type Config struct {
	Level    string `json:"level" yaml:"level"`
	Format   string `json:"format" yaml:"format"`
	InfoLog  string `json:"infoLog" yaml:"infoLog"`
	ErrorLog string `json:"errorLog" yaml:"errorLog"`
}

type Logger struct {
	logger *logrus.Logger
	hs     []HookCloser
	closed atomic.Bool
}

func (l *Logger) Close() error {
	if l.closed.Load() {
		return errors.New("logger is closed")
	}
	l.closed.Store(true)

	var err error
	for _, h := range l.hs {
		err = errors.Join(err, h.Close())
	}
	return err
}

func Trace(args ...any) {
	defaultLogger.logger.Trace(args...)
}

func Tracef(format string, args ...any) {
	defaultLogger.logger.Tracef(format, args...)
}

func Debug(args ...any) {
	defaultLogger.logger.Debug(args...)
}

func Debugf(format string, args ...any) {
	defaultLogger.logger.Debugf(format, args...)
}

func Info(args ...any) {
	defaultLogger.logger.Info(args...)
}

func Infof(format string, args ...any) {
	defaultLogger.logger.Infof(format, args...)
}

func Warn(args ...any) {
	defaultLogger.logger.Warn(args...)
}

func Warnf(format string, args ...any) {
	defaultLogger.logger.Warnf(format, args...)
}

func Error(args ...any) {
	defaultLogger.logger.Error(args...)
}

func Errorf(format string, args ...any) {
	defaultLogger.logger.Errorf(format, args...)
}

func Panic(args ...any) {
	defaultLogger.logger.Panic(args...)
}

func Panicf(format string, args ...any) {
	defaultLogger.logger.Panicf(format, args...)
}

func Fatal(args ...any) {
	defaultLogger.logger.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	defaultLogger.logger.Fatalf(format, args...)
}

func Close() error {
	return defaultLogger.Close()
}
