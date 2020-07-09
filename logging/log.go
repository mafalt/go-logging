package logging

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const logFlags = log.Ldate | log.Ltime | log.Lmicroseconds

// LogSeverity specifies possible logging severities:
// 10 FATAL
// 20 ERROR
// 30 WARNING
// 40 INFORMATION
// 50 DEBUG
// 60 VERBOSE
type LogSeverity uint16

const (
	// Fatal message
	Fatal LogSeverity = 10
	// Error message
	Error LogSeverity = 20
	// Warning message
	Warning LogSeverity = 30
	// Information message
	Information LogSeverity = 40
	// Debug message
	Debug LogSeverity = 50
	// Verbose message
	Verbose LogSeverity = 60
)

// LogType specifies logging target (file, screen, ...)
type LogType byte

const (
	// File target
	File LogType = iota + 1
	// Screen target
	Screen
)

// Logger type encapsulates work with raw logger to write log messages
type Logger struct {
	rawLogger *log.Logger
	severity  LogSeverity
	logType   LogType
	prefix    string
}

// Log implements ILog interface and provides logging functionality
type Log struct {
	loggers []*Logger
}

// ILog interface provides common interface for logging
type ILog interface {
	SetupLoggers(cfg LogConfig) error
	Fatal(msg string)
	Fatalf(msg string, args ...interface{})
	Error(msg string)
	Errorf(msg string, args ...interface{})
	Errore(err error)
	Warning(msg string)
	Warningf(msg string, args ...interface{})
	Info(msg string)
	Infof(msg string, args ...interface{})
	Debug(msg string)
	Debugf(msg string, args ...interface{})
	Verbose(msg string)
	Verbosef(msg string, args ...interface{})
}

// LogConfig type provides logging configuration
type LogConfig struct {
	Loggers []struct {
		LogType  string      `json:"logType" yaml:"logType"`
		Severity LogSeverity `json:"severity" yaml:"severity"`
		Rotate   bool        `json:"rotate" yaml:"rotate"`
		Path     string      `json:"path" yaml:"path"`
		Prefix   string      `json:"prefix" yaml:"prefix"`
	} `json:"logger" yaml:"loggers"`
}

var logStrings = []string{
	"FATAL  ",
	"ERROR  ",
	"WARNING",
	"INFO   ",
	"DEBUG  ",
	"VERBOSE",
}

func getLogTypeString(severity LogSeverity) string {
	return logStrings[severity/10-1]
}

func (l *Logger) logger() *log.Logger {
	return l.rawLogger
}

func (l *Log) createLogDir(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, os.ModeDir|os.ModePerm)
		}
	}

	return nil
}

func (l *Log) createLogFile(logFilePath string, rotate bool) (*os.File, error) {
	_, err := os.Stat(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Create(logFilePath)
		}

		return nil, err
	}

	if rotate {
		err := os.Rename(logFilePath, fmt.Sprintf("%s.%s", logFilePath, time.Now().Format("20060102150405")))
		if err != nil {
			return nil, err
		}
	}

	return os.Create(logFilePath)
}

// SetupLoggers method configures loggers to be used for logging
func (l *Log) SetupLoggers(cfg LogConfig) error {
	if cfg.Loggers == nil || len(cfg.Loggers) == 0 {
		return fmt.Errorf("unable to setup loggers")
	}

	for _, item := range cfg.Loggers {
		lg := &Logger{}
		switch strings.ToLower(item.LogType) {
		case "file":
			lg.logType = File
		case "screen":
			lg.logType = Screen
		default:
			return fmt.Errorf("%s is invalid log type", item.LogType)
		}
		lg.severity = LogSeverity(item.Severity)

		switch lg.logType {
		case Screen:
			lg.rawLogger = log.New(os.Stdout, item.Prefix, logFlags)
		case File:
			logDir := path.Dir(item.Path)
			if err := l.createLogDir(logDir); err != nil {
				return fmt.Errorf("failed to create logging directory: %s", err.Error())
			}

			f, err := l.createLogFile(item.Path, item.Rotate)
			if err != nil {
				return fmt.Errorf("failed to create log file: %s", err.Error())
			}

			lg.rawLogger = log.New(f, item.Prefix, logFlags)
		}

		l.loggers = append(l.loggers, lg)
	}

	return nil
}

func (l *Log) writeMessage(severity LogSeverity, msg string) {
	for _, lg := range l.loggers {
		if lg.severity >= severity {
			lg.logger().Printf("%s %s", getLogTypeString(severity), msg)
		}
	}
}

func (l *Log) writeMessagef(severity LogSeverity, msg string, args ...interface{}) {
	for _, lg := range l.loggers {
		if lg.severity >= severity {
			lg.logger().Printf(fmt.Sprintf("%s %s", getLogTypeString(severity), msg), args...)
		}
	}
}

// Fatal writes fatal message into the log
func (l *Log) Fatal(msg string) {
	l.writeMessage(Fatal, msg)
}

// Fatalf writes formatted fatal message into the log
func (l *Log) Fatalf(msg string, args ...interface{}) {
	l.writeMessagef(Fatal, msg, args...)
}

// Error writes error message into the log
func (l *Log) Error(msg string) {
	l.writeMessage(Error, msg)
}

// Errorf writes formatted error message into the log
func (l *Log) Errorf(msg string, args ...interface{}) {
	l.writeMessagef(Error, msg, args...)
}

// Errore writes error message into the log
func (l *Log) Errore(err error) {
	l.Error(err.Error())
}

// Warning writes warning message into the log
func (l *Log) Warning(msg string) {
	l.writeMessage(Warning, msg)
}

// Warningf writes formatted warning message into the log
func (l *Log) Warningf(msg string, args ...interface{}) {
	l.writeMessagef(Warning, msg, args...)
}

// Info writes informational message into the log
func (l *Log) Info(msg string) {
	l.writeMessage(Information, msg)
}

// Infof writes formatted informational message into the log
func (l *Log) Infof(msg string, args ...interface{}) {
	l.writeMessagef(Information, msg, args...)
}

// Debug writes debug message into the log
func (l *Log) Debug(msg string) {
	l.writeMessage(Debug, msg)
}

// Debugf writes formatted debug message into the log
func (l *Log) Debugf(msg string, args ...interface{}) {
	l.writeMessagef(Debug, msg, args...)
}

// Verbose writes verbose message into the log
func (l *Log) Verbose(msg string) {
	l.writeMessage(Verbose, msg)
}

// Verbosef writes formatted verbose message into the log
func (l *Log) Verbosef(msg string, args ...interface{}) {
	l.writeMessagef(Verbose, msg, args...)
}
