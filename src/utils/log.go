package utils

import (
	"log"
	"fmt"
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

func MajorError(message string, err error) {
	ll := LoggingLevelLabels[GetMainConfig().LoggingLevel]
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	if ll <= ERROR {
		log.Println(Red + "[ERROR] " + message + " : " + errStr + Reset)
	}
	
	TriggerEvent(
		"cosmos.error",
		"Critical Error",
		"error",
		"",
		map[string]interface{}{
			"message": message,
			"error": errStr,
	})

	WriteNotification(Notification{
		Recipient: "admin",
		Title: "header.notification.title.serverError",
		Message: message + " : " + errStr,
		Vars: "",
		Level: "error",
	})
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

func DoWarn(format string, a ...interface{}) string {
	message := fmt.Sprintf(format, a...)
	// \033[1;33m is the ANSI code for bold yellow
	// \033[0m resets the color
	return fmt.Sprintf("\033[1;33m[WARN] %s\033[0m", message)
}

func DoErr(format string, a ...interface{}) string {
	message := fmt.Sprintf(format, a...)
	// \033[1;31m is the ANSI code for bold red
	// \033[0m resets the color
	return fmt.Sprintf("\033[1;31m[ERROR] %s\033[0m", message)
}

func DoSuccess(format string, a ...interface{}) string {
	message := fmt.Sprintf(format, a...)
	// \033[1;32m is the ANSI code for bold green
	// \033[0m resets the color
	return fmt.Sprintf("\033[1;32m[SUCCESS] %s\033[0m", message)
}