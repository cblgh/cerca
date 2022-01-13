package logger

import (
	"log"
)

type LogLevel uint32

const (
	LevelFatal LogLevel = iota
	LevelError
	LevelInfo
	LevelDebug
)

var globalLevel = LevelInfo

func SetLevel(level LogLevel) {
	globalLevel = level
}

func Fatal(format string, v ...interface{}) {
	if globalLevel >= LevelFatal {
		log.Fatalf("[FATAL] "+format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if globalLevel >= LevelError {
		log.Printf("[ERROR] "+format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if globalLevel >= LevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

func Debug(format string, v ...interface{}) {
	if globalLevel >= LevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}
