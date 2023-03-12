package utils

import (
	"log"
)

var Reset  = "\033[0m"
var Red    = "\033[31m"
var Green  = "\033[32m"
var Yellow = "\033[33m"
var Blue   = "\033[34m"
var Purple = "\033[35m"
var Cyan   = "\033[36m"
var Gray   = "\033[37m"
var White  = "\033[97m"

func Debug(message string) {
	ll := LoggingLevelLabels[GetMainConfig().LoggingLevel]
	if ll <= DEBUG {
		log.Println(Purple + "[DEBUG] " + message + Reset)
	}
}

func Log(message string) {
	ll := LoggingLevelLabels[GetMainConfig().LoggingLevel]
	if ll <= INFO {
		log.Println(Blue + "[INFO] " + message + Reset)
	}
}

func Warn(message string) {
	ll := LoggingLevelLabels[GetMainConfig().LoggingLevel]
	if ll <= WARNING {
		log.Println(Yellow + "[WARN] " + message + Reset)
	}
}

func Error(message string, err error) {
	ll := LoggingLevelLabels[GetMainConfig().LoggingLevel]
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	if ll <= ERROR {
		log.Println(Red + "[ERROR] " + message + " : " + errStr + Reset)
	}
}

func Fatal(message string, err error) {
	ll := LoggingLevelLabels[GetMainConfig().LoggingLevel]
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	if ll <= ERROR {
		log.Fatal(Red + "[FATAL] " + message + " : " + errStr + Reset)
	}
}
