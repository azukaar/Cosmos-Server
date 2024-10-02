package utils

import (
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Reset    = "\033[0m"
	Bold		 = "\033[1m"
	nRed     = "\033[31m"
	nGreen   = "\033[32m"
	nYellow  = "\033[33m"
	nBlue    = "\033[34m"
	nPurple  = "\033[35m"
	nCyan    = "\033[36m"
	nGray    = "\033[37m"
	nWhite   = "\033[97m"
	nMagenta = "\033[95m"
	nBlack   = "\033[30m"
	bRed     = "\033[41m"
	bGreen   = "\033[42m"
	bYellow  = "\033[43m"
	bBlue    = "\033[44m"
	bMagenta = "\033[45m"
	bCyan    = "\033[46m"
	bGray    = "\033[47m"
	bWhite   = "\033[107m"
	bPurple  = "\033[45m"
)

type LogLevel int

var (
	logger      *log.Logger
	errorLogger *log.Logger
)

func InitLogs() {
	// Set up lumberjack log rotation
	ljLogger := &lumberjack.Logger{
		Filename:   CONFIGFOLDER + "cosmos.log",
		MaxSize:    15, // megabytes
		MaxBackups: 2,
		MaxAge:     16, // days
		Compress:   true,
	}

	// Create a multi-writer to log to both file and stdout
	multiWriter := io.MultiWriter(ljLogger, os.Stdout)

	// Create a multi-writer for errors to log to both file and stderr
	errorWriter := io.MultiWriter(ljLogger, os.Stderr)

	// Create loggers
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime)
	errorLogger = log.New(errorWriter, "", log.Ldate|log.Ltime)

	log.Println("Logging initialized") // shows
	logger.Println("Logging initialized") // does not show
}

func logMessage(level LogLevel, prefix, prefixColor, color, message string) {
	ll := LoggingLevelLabels[GetMainConfig().LoggingLevel]
	if ll <= level {
		logString := prefixColor + Bold + prefix + Reset + " " + color + message + Reset
		if level >= ERROR {
			errorLogger.Println(logString)
		} else {
			logger.Println(logString)
		}
	}
}

func Debug(message string) {
	logMessage(DEBUG, "[DEBUG]", bPurple, nPurple, message)
}

func Log(message string) {
	logMessage(INFO, "[INFO]", bBlue, nBlue, message)
}

func LogReq(message string) {
	logMessage(INFO, "[REQ]", bGreen, nGreen, message)
}

func Warn(message string) {
	logMessage(WARNING, "[WARN]", bYellow, nYellow, message)
}

func Error(message string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	logMessage(ERROR, "[ERROR]", bRed, nRed, message+" : "+errStr)
}

func MajorError(message string, err error) {
	Error(message, err)

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	TriggerEvent(
		"cosmos.error",
		"Critical Error",
		"error",
		"",
		map[string]interface{}{
			"message": message,
			"error":   errStr,
		})

	WriteNotification(Notification{
		Recipient: "admin",
		Title:     "header.notification.title.serverError",
		Message:   message + " : " + errStr,
		Vars:      "",
		Level:     "error",
	})
}

func Fatal(message string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	errorLogger.Fatal(bRed + "[FATAL]" + Reset + " " + nRed + message + " : " + errStr + Reset)
}

func DoWarn(format string, a ...interface{}) string {
	message := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%s%s[WARN]%s %s%s%s", bYellow, nBlack, Reset, nYellow, message, Reset)
}

func DoErr(format string, a ...interface{}) string {
	message := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%s%s[ERROR]%s %s%s%s", bRed, nWhite, Reset, nRed, message, Reset)
}

func DoSuccess(format string, a ...interface{}) string {
	message := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%s%s[SUCCESS]%s %s%s%s", bGreen, nBlack, Reset, nGreen, message, Reset)
}