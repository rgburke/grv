package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	logLogrusRepo     = "github.com/Sirupsen/logrus"
	logFileDateFormat = "2006-01-02 15:04:05.000-0700"
)

var logFile string

type logFormatter struct{}

func (formatter logFormatter) Format(entry *log.Entry) ([]byte, error) {
	var buffer bytes.Buffer
	var file string

	if entry.HasCaller() {
		file = fmt.Sprintf("%v:%v", path.Base(entry.Caller.File), entry.Caller.Line)
	} else {
		file = "Unknown"
	}

	formatter.formatBracketEntry(&buffer, entry.Time.Format(logFileDateFormat))
	formatter.formatBracketEntry(&buffer, strings.ToUpper(entry.Level.String()))
	formatter.formatBracketEntry(&buffer, file)

	buffer.WriteString("- ")

	for _, char := range entry.Message {
		switch {
		case char == '\n':
			buffer.WriteString("\\n")
		case char < 32 || char == 127:
			buffer.WriteString(NonPrintableCharString(char))
		default:
			buffer.WriteRune(char)
		}
	}

	buffer.WriteRune('\n')

	return buffer.Bytes(), nil
}

func (formatter logFormatter) formatBracketEntry(buffer *bytes.Buffer, value string) {
	buffer.WriteRune('[')
	buffer.WriteString(value)
	buffer.WriteString("] ")
}

// LogFile returns the path of the file GRV is logging to
func LogFile() string {
	return logFile
}

// InitialiseLogging sets up logging
func InitialiseLogging(logLevel, logFilePath string) {
	if logLevel == MnLogLevelDefault {
		log.SetOutput(ioutil.Discard)
		return
	}

	logFile = logFilePath

	logLevels := map[string]log.Level{
		"PANIC": log.PanicLevel,
		"FATAL": log.FatalLevel,
		"ERROR": log.ErrorLevel,
		"WARN":  log.WarnLevel,
		"INFO":  log.InfoLevel,
		"DEBUG": log.DebugLevel,
		"TRACE": log.TraceLevel,
	}

	if level, ok := logLevels[logLevel]; ok {
		log.SetLevel(level)
	} else {
		log.Fatalf("Invalid logLevel: %v", logLevel)
	}

	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Unable to open log file %v for writing: %v", logFilePath, err)
	}

	log.SetOutput(file)

	log.SetReportCaller(true)

	log.SetFormatter(logFormatter{})
}
