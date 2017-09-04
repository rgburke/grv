package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

const (
	mnRepoFilePathDefault = "."
	mnLogFilePathDefault  = "grv.log"
	// MnLogLevelDefault is the default log level for grv
	MnLogLevelDefault = "NONE"
)

type grvArgs struct {
	repoFilePath string
	logLevel     string
	logFilePath  string
}

func main() {
	args := parseArgs()
	InitialiseLogging(args.logLevel, args.logFilePath)

	grv := NewGRV()

	if err := grv.Initialise(args.repoFilePath); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Unable to initialise grv: %v\n", err)
		grv.Free()
		log.Fatal(err)
	}

	grv.Run()

	grv.Free()

	log.Info("Exiting normally")
}

func parseArgs() *grvArgs {
	repoFilePathPtr := flag.String("repoFilePath", mnRepoFilePathDefault, "Repository file path")
	logLevelPtr := flag.String("logLevel", MnLogLevelDefault, "Logging level [NONE|PANIC|FATAL|ERROR|WARN|INFO|DEBUG]")
	logFilePathPtr := flag.String("logFile", mnLogFilePathDefault, "Log file path")

	flag.Parse()

	return &grvArgs{
		repoFilePath: *repoFilePathPtr,
		logLevel:     *logLevelPtr,
		logFilePath:  *logFilePathPtr,
	}
}
