package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	mnRepoFilePathDefault     = "."
	mnWorkTreeFilePathDefault = ""
	mnLogFilePathDefault      = "grv.log"
	// MnGenerateDocumentationEnv is set when documentation should be generated
	MnGenerateDocumentationEnv = "GRV_GENERATE_DOCUMENTATION"
	// MnLogLevelDefault is the default log level for grv
	MnLogLevelDefault = "NONE"
)

var (
	version       = "Unknown"
	buildDateTime = "Unknown"
)

type grvArgs struct {
	repoFilePath     string
	workTreeFilePath string
	logLevel         string
	logFilePath      string
	version          bool
	readOnly         bool
}

func main() {
	args := parseArgs()
	if args.version {
		printVersion()
		return
	}

	if os.Getenv(MnGenerateDocumentationEnv) != "" {
		if err := GenerateDocumentation(); err != nil {
			log.Fatalf("Failed to generate documentation: %v", err)
		} else {
			fmt.Printf("Generated documentation\n")
		}

		return
	}

	InitialiseLogging(args.logLevel, args.logFilePath)
	log.Info(getVersion())

	log.Debugf("Creating GRV instance")
	grv := NewGRV(args.readOnly)

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
	logLevelPtr := flag.String("logLevel", MnLogLevelDefault, "Logging level [NONE|PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE]")
	logFilePathPtr := flag.String("logFile", mnLogFilePathDefault, "Log file path")
	versionPtr := flag.Bool("version", false, "Print version")
	readOnlyPtr := flag.Bool("readOnly", false, "Run grv in read only mode")

	flag.Parse()

	return &grvArgs{
		repoFilePath:     *repoFilePathPtr,
		workTreeFilePath: *workTreeFilePathPtr,
		logLevel:         *logLevelPtr,
		logFilePath:      *logFilePathPtr,
		version:          *versionPtr,
		readOnly:         *readOnlyPtr,
	}
}

func getVersion() string {
	return fmt.Sprintf("GRV - Git Repository Viewer %v (compiled with %v at %v)", version, runtime.Version(), buildDateTime)
}

func printVersion() {
	fmt.Printf("%v\n", getVersion())
}

// GenerateCommandLineArgumentsHelpSections generates help documentation for command line arguments
func GenerateCommandLineArgumentsHelpSections() *HelpSection {
	description := []HelpSectionText{
		{text: "GRV accepts the following command line arguments:"},
		{},
	}

	var buffer bytes.Buffer
	flag.CommandLine.SetOutput(&buffer)
	flag.CommandLine.PrintDefaults()

	scanner := bufio.NewScanner(&buffer)
	for scanner.Scan() {
		description = append(description, HelpSectionText{
			text:             strings.TrimLeft(scanner.Text(), " "),
			themeComponentID: CmpHelpViewSectionCodeBlock,
		})
	}

	return &HelpSection{
		title:       HelpSectionText{text: "Command Line Arguments"},
		description: description,
	}
}
