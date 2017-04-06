package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	LOG_LOGRUS_REPO      = "github.com/Sirupsen/logrus"
	LOG_FILE_DATE_FORMAT = "2006-01-02 15:04:05.000-0700"
)

type FileHook struct{}

func (fileHook FileHook) Fire(entry *log.Entry) (err error) {
	pc := make([]uintptr, 5, 5)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()

		if !strings.Contains(name, LOG_LOGRUS_REPO) {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data["file"] = fmt.Sprintf("%v:%v", path.Base(file), line)
			break
		}
	}

	return
}

func (fileHook FileHook) Levels() []log.Level {
	return log.AllLevels
}

type LogFormatter struct{}

func (logFormatter LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	var buffer bytes.Buffer
	file, _ := entry.Data["file"].(string)

	logFormatter.formatBracketEntry(&buffer, entry.Time.Format(LOG_FILE_DATE_FORMAT))
	logFormatter.formatBracketEntry(&buffer, strings.ToUpper(entry.Level.String()))
	logFormatter.formatBracketEntry(&buffer, file)

	buffer.WriteString("- ")
	buffer.WriteString(entry.Message)
	buffer.WriteRune('\n')

	return buffer.Bytes(), nil
}

func (logFormatter LogFormatter) formatBracketEntry(buffer *bytes.Buffer, value string) {
	buffer.WriteRune('[')
	buffer.WriteString(value)
	buffer.WriteString("] ")
}

func InitialiseLogging(logLevel, logFilePath string) {
	if logLevel == MN_LOG_LEVEL_DEFAULT {
		log.SetOutput(ioutil.Discard)
		return
	}

	logLevels := map[string]log.Level{
		"PANIC": log.PanicLevel,
		"FATAL": log.FatalLevel,
		"ERROR": log.ErrorLevel,
		"WARN":  log.WarnLevel,
		"INFO":  log.InfoLevel,
		"DEBUG": log.DebugLevel,
	}

	if level, ok := logLevels[logLevel]; ok {
		log.SetLevel(level)
	} else {
		log.Fatalf("Invalid logLevel: %v", logLevel)
	}

	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("Unable to open log file %v for writing: %v", logFilePath, err)
	}

	log.SetOutput(file)

	log.SetFormatter(LogFormatter{})

	log.AddHook(FileHook{})
}
