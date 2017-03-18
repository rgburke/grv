package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

const (
	MN_REPO_FILE_PATH_DEFAULT = "."
	MN_LOG_LEVEL_DEFAULT      = "NONE"
	MN_LOG_FILE_PATH_DEFAULT  = "grv.log"
	MN_LOG_FILE_DATE_FORMAT   = "2006-01-02T15:04:05.000-0700"
)

type GRVArgs struct {
	repoFilePath string
	logLevel     string
	logFilePath  string
}

func main() {
	args := parseArgs()
	initialiseLogging(args.logLevel, args.logFilePath)

	grv := NewGRV()

	if err := grv.Initialise(args.repoFilePath); err != nil {
		log.Fatal(err)
	}

	grv.Run()

	grv.Free()

	log.Info("Exiting normally")
}

func parseArgs() *GRVArgs {
	repoFilePathPtr := flag.String("repoFilePath", MN_REPO_FILE_PATH_DEFAULT, "Repository file path")
	logLevelPtr := flag.String("logLevel", MN_LOG_LEVEL_DEFAULT, "Logging level [NONE|PANIC|FATAL|ERROR|WARN|INFO|DEBUG]")
	logFilePathPtr := flag.String("logFile", MN_LOG_FILE_PATH_DEFAULT, "Log file path")

	flag.Parse()

	return &GRVArgs{
		repoFilePath: *repoFilePathPtr,
		logLevel:     *logLevelPtr,
		logFilePath:  *logFilePathPtr,
	}
}

func initialiseLogging(logLevel, logFilePath string) {
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

	formatter := &log.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: MN_LOG_FILE_DATE_FORMAT,
	}

	log.SetFormatter(formatter)
}
