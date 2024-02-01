package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Format string
type Level string

const (
	timeFormat = time.RFC3339Nano

	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
	Fatal Level = "fatal"
	Panic Level = "panic"

	JSON  Format = "json"
	Plain Format = "plain"
)

var (
	Version     = "v0.1.0" // Semantic version
	FullVersion = fmt.Sprintf("%s-%v", Version, GitCommitHash[0:8])
)

// actionFormatter duplicate the log message inside a field action
// this is useful for log analysis.
type actionFormatter struct {
	original logrus.Formatter
}

type ddErr struct {
	Message any `json:"message"`
}

// errorReMapper remap the field error value inside struct
// error fields conflicts with error.kind
type errorReMapper struct {
	logrus.Formatter
}

// NewLogger initializes a new (logrus) Logger instance
// Supported log types are: text, json, none
func NewLogger(logLevel Level, logFormat Format, tofile bool, service string) (*logrus.Entry, error) {
	if err := checkLevel(logLevel, service); err != nil {
		return nil, err
	}
	if err := checkFormat(logFormat, service); err != nil {
		return nil, err
	}
	debugWarnMsg := fmt.Sprintf("%s RUNNING IN DEBUG MODE. DO NOT RUN IN PRODUCTION ENVIRONMENT", strings.ToUpper(service))
	// Set logging to json
	l := newFormattedLogger(logLevel, logFormat, debugWarnMsg)

	// OPTIONAL OUTPUT TO .log file (Default false)
	if tofile {
		f, err := os.OpenFile(fmt.Sprintf("%s.log", service), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, fmt.Errorf("error opening file: %v", err)
		}

		// assign it to the standard logger
		l.SetOutput(f)
		l.Infof("Logging to file: %v", fmt.Sprintf("%s.log", service))
	}

	logger := logrus.NewEntry(l)
	logger.Level = l.Level
	logger = logger.WithFields(logrus.Fields{
		"serviceName": service,
		"version":     FullVersion,
	})
	return logger, nil
}

// return a formatted Logger object (log format is JSON by defulat)
func newFormattedLogger(logLevel Level, logFormat Format, debugWarnMsg string) *logrus.Logger {
	l := logrus.New()
	var formatter logrus.Formatter
	// Select log Format
	switch logFormat {
	case Plain:
		formatter = &logrus.TextFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyMsg: "message",
			},
			TimestampFormat: timeFormat,
		}
	default:
		formatter = &errorReMapper{&actionFormatter{&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyMsg: "message",
			},
			TimestampFormat: timeFormat,
		}}}
	}

	l.SetFormatter(formatter)
	level := logLevelFromString(logLevel)
	l.Level = level
	if logLevel == Debug {
		l.Warn(debugWarnMsg)
	}
	return l
}

// Format serializes the entry into a byte slice using the logrus formatter
func (formatter *actionFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return formatter.original.Format(entry)
}

func (formatter *errorReMapper) Format(entry *logrus.Entry) ([]byte, error) {
	errorMessage := entry.Data["error"]
	if errorMessage != nil {
		entry.Data["error"] = ddErr{
			Message: errorMessage,
		}
	}
	return formatter.Formatter.Format(entry)
}

// Validation

func checkLevel(w Level, service string) error {
	switch w {
	case Debug, Info, Warn, Error, Fatal, Panic:
		return nil
	default:
		return fmt.Errorf("invalid %s log level input '%v'", service, w)
	}
}

func checkFormat(w Format, service string) error {
	switch w {
	case JSON, Plain:
		return nil
	default:
		return fmt.Errorf("invalid %s log format input '%v'", service, w)
	}
}

// logLevelFromString returns log level (Int) from string input
// Log Level (error, info, debug) from string input
func logLevelFromString(logLevel Level) logrus.Level {
	switch logLevel {
	case Panic:
		return logrus.PanicLevel
	case Fatal:
		return logrus.FatalLevel
	case Error:
		return logrus.ErrorLevel
	case Warn:
		return logrus.WarnLevel
	case Info:
		return logrus.InfoLevel
	case Debug:
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}
