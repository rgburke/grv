package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
)

const (
	MN_REPO_FILE_PATH_DEFAULT = "."
	MN_LOG_LEVEL_DEFAULT      = "NONE"
	MN_LOG_FILE_PATH_DEFAULT  = "grv.log"
)

type GRVArgs struct {
	repoFilePath string
	logLevel     string
	logFilePath  string
}

func main() {
	args := parseArgs()
	InitialiseLogging(args.logLevel, args.logFilePath)

	grv := NewGRV()

	if err := grv.Initialise(args.repoFilePath); err != nil {
		grv.Free()
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
