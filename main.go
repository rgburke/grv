package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

const (
	NO_LOGGING = "NONE"
	LOG_FILE   = "grv.log"
)

func main() {
	parseArgs()

	grv := NewGRV()

	if err := grv.Initialise("."); err != nil {
		log.Fatal(err)
	}

	grv.Run()

	grv.Free()
}

func parseArgs() {
	logLevelPtr := flag.String("logLevel", NO_LOGGING, "Sets the logging level [NONE|PANIC|FATAL|ERROR|WARN|INFO|DEBUG]")
	logFilePtr := flag.String("logFile", LOG_FILE, "Specifies the log file path")

	flag.Parse()

	initialiseLogging(*logLevelPtr, *logFilePtr)
}

func initialiseLogging(logLevel, logFile string) {
	if logLevel == NO_LOGGING {
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

	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("Unable to open log file %v for writing: %v", logFile, err)
	}

	log.SetOutput(file)

	formatter := &log.TextFormatter{}
	formatter.DisableColors = true
	log.SetFormatter(formatter)
}
