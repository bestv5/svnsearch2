package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

type Logger struct {
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	logFile       *os.File
	logLevel      LogLevel
}

var defaultLogger *Logger

func InitLogger(logDir string, logLevel string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	logFileName := fmt.Sprintf("svnsearch_%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	defaultLogger = &Logger{
		debugLogger:   log.New(multiWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:    log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger: log.New(multiWriter, "[WARNING] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:   log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		logFile:       logFile,
		logLevel:      parseLogLevel(logLevel),
	}

	return nil
}

func parseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.logLevel <= DEBUG {
		l.debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.logLevel <= INFO {
		l.infoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warning(format string, v ...interface{}) {
	if l.logLevel <= WARNING {
		l.warningLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.logLevel <= ERROR {
		l.errorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func Debug(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, v...)
	}
}

func Warning(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warning(format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, v...)
	}
}

func Close() error {
	if defaultLogger != nil {
		return defaultLogger.Close()
	}
	return nil
}
