package main

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// CustomFormatter реализует интерфейс logrus.Formatter
type CustomFormatter struct{}

// Format определяет, как будет выводиться лог
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Добавляем префикс [DRYRUN], если флаг установлен
	prefix := ""
	//if DRY_RUN_FLAG {
	//	prefix = "[DRYRUN] "
	//}

	// Строковый формат, который будет выводить время, уровень и сообщение
	return []byte(time.Now().Format("2006-01-02 15:04:05") + " " + prefix + "[" + entry.Level.String() + "] " + entry.Message + "\n"), nil
}

func setupLogging(loggingConfig LoggingConfig) {
	/*
		if loggingConfig == nil {
			fmt.Println("No logging configuration provided. Using default settings.")
			logrus.SetLevel(logrus.InfoLevel)
			logrus.SetFormatter(&logrus.TextFormatter{})
			return
		}
	*/
	// Устанавливаем пользовательский формат, если он указан
	if loggingConfig.Formatter == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else if loggingConfig.Formatter == "default" {
		logrus.SetFormatter(&CustomFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	// Устанавливаем уровень логирования
	switch loggingConfig.Level {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	default:
		fmt.Printf("Unknown logging level '%s', defaulting to INFO\n", loggingConfig.Level)
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// Функция для проверки уровня логирования
func shouldLog(level string, logLevel string) bool {
	switch logLevel {
	case "DEBUG":
		return true
	case "INFO":
		return level == "INFO" || level == "ERROR"
	case "ERROR":
		return level == "ERROR"
	default:
		return false
	}
}

// Функция для вывода сообщений в консоль в зависимости от уровня логирования
func logMessage(level string, format string, args ...interface{}) {
	if shouldLog(level, level) {
		// Форматируем строку с помощью fmt.Sprintf
		message := fmt.Sprintf(format, args...)

		switch level {
		case "DEBUG":
			logrus.Debug(message)
		case "INFO":
			logrus.Info(message)
		case "ERROR":
			logrus.Error(message)
		}
	}
}
