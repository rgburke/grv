package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

const (
	mnRepoFilePathDefault     = "."
	mnWorkTreeFilePathDefault = ""
	mnLogFilePathDefault      = "grv.log"
	// MnLogLevelDefault is the default log level for grv
	MnLogLevelDefault = "NONE"
)

var (
	version       = "Unknown"
	headOid       = "Unknown"
	buildDateTime = "Unknown"
)

type grvArgs struct {
	repoFilePath     string
	workTreeFilePath string
	logLevel         string
	logFilePath      string
	version          bool
}

func main() {
	args := parseArgs()
	if args.version {
		printVersion()
		return
	}

	InitialiseLogging(args.logLevel, args.logFilePath)
	log.Info(getVersion())

	log.Debugf("Creating GRV instance")
	grv := NewGRV()

	if err := grv.Initialise(args.repoFilePath, args.workTreeFilePath); err != nil {
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
	workTreeFilePathPtr := flag.String("workTreeFilePath", mnWorkTreeFilePathDefault, "Work tree file path")
	logLevelPtr := flag.String("logLevel", MnLogLevelDefault, "Logging level [NONE|PANIC|FATAL|ERROR|WARN|INFO|DEBUG]")
	logFilePathPtr := flag.String("logFile", mnLogFilePathDefault, "Log file path")
	versionPtr := flag.Bool("version", false, "Print version")

	flag.Parse()

	return &grvArgs{
		repoFilePath:     *repoFilePathPtr,
		workTreeFilePath: *workTreeFilePathPtr,
		logLevel:         *logLevelPtr,
		logFilePath:      *logFilePathPtr,
		version:          *versionPtr,
	}
}

func getVersion() string {
	return fmt.Sprintf("GRV - Git Repository Viewer %v (commit: %v, compiled: %v)", version, headOid, buildDateTime)
}

func printVersion() {
	fmt.Printf("%v\n", getVersion())
}
