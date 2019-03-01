// GRV is a terminal interface for viewing git repositories
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	fs "github.com/rjeczalik/notify"
)

const (
	grvInputBufferSize       = 100
	grvActionBufferSize      = 100
	grvEventBufferSize       = 100
	grvErrorBufferSize       = 100
	grvDisplayBufferSize     = 50
	grvMaxDrawFrequency      = time.Millisecond * 50
	grvMinErrorDisplay       = time.Second * 2
	grvMaxGitStatusFrequency = time.Millisecond * 500
)

type gRVChannels struct {
	exitCh     chan bool
	inputKeyCh chan string
	actionCh   chan Action
	eventCh    chan Event
	displayCh  chan bool
	errorCh    chan error
}

func (grvChannels gRVChannels) Channels() *channels {
	return &channels{
		displayCh:  grvChannels.displayCh,
		exitCh:     grvChannels.exitCh,
		errorCh:    grvChannels.errorCh,
		actionCh:   grvChannels.actionCh,
		eventCh:    grvChannels.eventCh,
		inputKeyCh: grvChannels.inputKeyCh,
	}
}

// Channels contains channels used for communication within grv
type Channels interface {
	UpdateDisplay()
	Exit() bool
	ReportError(err error)
	ReportErrors(errors []error)
	DoAction(action Action)
	ReportEvent(event Event)
	ReportStatus(format string, args ...interface{})
}

type channels struct {
	displayCh  chan<- bool
	exitCh     <-chan bool
	errorCh    chan<- error
	actionCh   chan<- Action
	eventCh    chan<- Event
	inputKeyCh chan<- string
}

// InputConsumer can consumer and process key string input
type InputConsumer interface {
	ProcessInput(input string)
}

// EventType identifies a type of event
type EventType int

// The event types available
const (
	NoEvent EventType = iota
	ViewRemovedEvent
)

// Event contains data that describes the reported event
type Event struct {
	EventType EventType
	Args      []interface{}
}

// EventListener is an entity capable of receiving events
type EventListener interface {
	HandleEvent(event Event) error
}

// GRV is the top level structure containing all state in the program
type GRV struct {
	repoInitialiser *RepositoryInitialiser
	repoData        *RepositoryData
	repoController  RepoController
	view            *View
	ui              UI
	channels        gRVChannels
	config          *Configuration
	inputBuffer     *InputBuffer
	input           *InputKeyMapper
	eventListeners  []EventListener
	variables       *GRVVariables
}

// UpdateDisplay sends a request to update the display
func (channels *channels) UpdateDisplay() {
	select {
	case channels.displayCh <- true:
	default:
	}
}

// Exit returns true if GRV is in the process of exiting
func (channels *channels) Exit() bool {
	select {
	case _, ok := <-channels.exitCh:
		return !ok
	default:
		return false
	}
}

// ReportError reports an error to be displayed
func (channels *channels) ReportError(err error) {
	if err != nil {
		select {
		case channels.errorCh <- err:
		default:
			log.Errorf("Unable to report error %v", err)
		}
	}
}

// ReportErrors reports multiple errors to be displayed
func (channels *channels) ReportErrors(errors []error) {
	for _, err := range errors {
		channels.ReportError(err)
	}
}

// DoAction sends an action to be executed
func (channels *channels) DoAction(action Action) {
	if action.ActionType != ActionNone {
		channels.actionCh <- action
	}
}

// ReportEvent sends the event to all listeners
func (channels *channels) ReportEvent(event Event) {
	if event.EventType != NoEvent {
		channels.eventCh <- event
	}
}

// ReportStatus updates the status bar with the provided status
func (channels *channels) ReportStatus(format string, args ...interface{}) {
	status := fmt.Sprintf(format, args...)

	if status != "" {
		channels.DoAction(Action{
			ActionType: ActionShowStatus,
			Args:       []interface{}{status},
		})
	}
}

// ProcessInput sends the provided input to be processed
func (channels *channels) ProcessInput(input string) {
	select {
	case channels.inputKeyCh <- input:
	default:
		log.Errorf("Unable to add input \"%v\" to input channel", input)
	}
}

// NewGRV creates a new instace of GRV
func NewGRV(readOnly bool) *GRV {
	grvChannels := gRVChannels{
		exitCh:     make(chan bool),
		inputKeyCh: make(chan string, grvInputBufferSize),
		actionCh:   make(chan Action, grvActionBufferSize),
		eventCh:    make(chan Event, grvEventBufferSize),
		displayCh:  make(chan bool, grvDisplayBufferSize),
		errorCh:    make(chan error, grvErrorBufferSize),
	}

	channels := grvChannels.Channels()
	keyBindings := NewKeyBindingManager()
	variables := NewGRVVariables()
	config := NewConfiguration(keyBindings, channels, variables, channels)

	repoDataLoader := NewRepoDataLoader(channels, config)
	repoData := NewRepositoryData(repoDataLoader, channels, variables)

	var repoController RepoController
	if readOnly {
		log.Info("Running grv in read only mode")
		repoController = NewReadOnlyRepositoryController()
	} else {
		repoController = NewGitCommandRepoController(repoData, channels, config)
	}

	ui := NewNCursesDisplay(channels, config)
	view := NewView(repoData, repoController, channels, config, variables)

	return &GRV{
		repoInitialiser: NewRepositoryInitialiser(),
		repoData:        repoData,
		repoController:  repoController,
		view:            view,
		ui:              ui,
		channels:        grvChannels,
		config:          config,
		inputBuffer:     NewInputBuffer(keyBindings),
		input:           NewInputKeyMapper(ui),
		eventListeners:  []EventListener{view, repoData, config},
		variables:       variables,
	}
}

// Initialise sets up all the components of GRV
func (grv *GRV) Initialise(repoPath, workTreePath string) (err error) {
	log.Info("Initialising GRV")

	channels := grv.channels.Channels()

	if configErrors := grv.config.Initialise(); configErrors != nil {
		channels.ReportErrors(configErrors)
	}

	if err = grv.repoInitialiser.CreateRepositoryInstance(repoPath, workTreePath); err != nil {
		return
	}

	if err = grv.repoData.Initialise(grv.repoInitialiser); err != nil {
		return
	}

	grv.repoController.Initialise(grv.repoInitialiser)

	if err = grv.ui.Initialise(); err != nil {
		return
	}

	if err = grv.view.Initialise(); err != nil {
		return
	}

	InitReadLine(channels, grv.config)

	return
}

// Free closes and frees any resources used by GRV
func (grv *GRV) Free() {
	log.Info("Freeing GRV")

	FreeReadLine()
	grv.ui.Free()
	grv.repoData.Free()
	grv.repoInitialiser.Free()
}

// Suspend prepares GRV to be suspended and sends a SIGTSTP
// to every process in the process group
func (grv *GRV) Suspend() {
	log.Info("Suspending GRV")

	grv.ui.Suspend()
	if err := syscall.Kill(0, syscall.SIGTSTP); err != nil {
		log.Errorf("Kill syscall failed. Error when attempting to suspend GRV: %v", err)
	}
}

// Resume is called on receipt of a SIGCONT and reinitialises the UI
func (grv *GRV) Resume() {
	log.Info("Resuming GRV")

	if err := grv.ui.Resume(); err != nil {
		log.Errorf("Error when attempting to resume GRV: %v", err)
	}

	grv.channels.displayCh <- true
}

// End signals GRV to stop
func (grv *GRV) End() {
	log.Info("Stopping GRV")

	close(grv.channels.exitCh)

	if err := grv.ui.CancelGetInput(); err != nil {
		log.Errorf("Error calling CancelGetInput: %v", err)
	}

	grv.view.Dispose()
}

// Run sets up the input, display, action and singal handler loops
// This function blocks until Exit is called
func (grv *GRV) Run() {
	var waitGroup sync.WaitGroup
	channels := grv.channels

	waitGroup.Add(1)
	go grv.runInputLoop(&waitGroup, channels.exitCh, channels.inputKeyCh, channels.errorCh)
	waitGroup.Add(1)
	go grv.runDisplayLoop(&waitGroup, channels.exitCh, channels.displayCh, channels.errorCh)
	waitGroup.Add(1)
	go grv.runHandlerLoop(&waitGroup, channels.exitCh, channels.inputKeyCh, channels.actionCh, channels.errorCh, channels.eventCh)
	waitGroup.Add(1)
	go grv.runSignalHandlerLoop(&waitGroup, channels.exitCh)
	waitGroup.Add(1)
	go grv.runFileSystemMonitorLoop(&waitGroup, channels.exitCh)

	channels.displayCh <- true

	log.Info("Waiting for loops to finish")
	waitGroup.Wait()
	log.Info("All loops finished")
}

func (grv *GRV) runInputLoop(waitGroup *sync.WaitGroup, exitCh chan bool, inputKeyCh chan<- string, errorCh chan<- error) {
	defer waitGroup.Done()
	defer log.Info("Input loop stopping")
	log.Info("Starting input loop")

	for {
		key, err := grv.input.GetKeyInput()
		if err != nil {
			errorCh <- err
		} else if key == "<Mouse>" {
			if mouseEvent, exists := grv.ui.GetMouseEvent(); exists {
				mouseEventAction, err := MouseEventAction(mouseEvent)
				if err != nil {
					errorCh <- err
				} else {
					grv.channels.actionCh <- mouseEventAction
				}
			}
		} else if key != "" {
			log.Debugf("Received keypress from UI %v", key)

			select {
			case inputKeyCh <- key:
			default:
				log.Errorf("Unable to add keypress %v to input channel", key)
			}
		}

		select {
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		default:
		}
	}
}

func (grv *GRV) runDisplayLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, displayCh <-chan bool, errorCh chan error) {
	defer waitGroup.Done()
	defer log.Info("Display loop stopping")
	log.Info("Starting display loop")

	var errors []error
	lastErrorReceivedTime := time.Now()
	channels := &channels{errorCh: errorCh}

	timer := time.NewTimer(time.Hour)
	timer.Stop()
	timerActive := false

	for {
		select {
		case <-displayCh:
			log.Debug("Received display refresh request")

			if !timerActive {
				timer.Reset(grvMaxDrawFrequency)
				timerActive = true
			}
		case <-timer.C:
			timerActive = false

			if lastErrorReceivedTime.Before(time.Now().Add(-grvMinErrorDisplay)) {
				errors = nil
			} else if errors != nil {
				grv.view.SetErrors(errors)
			}

			log.Debug("Refreshing display - Display refresh request received since last check")

			viewDimension := grv.ui.ViewDimension()

			wins, err := grv.view.Render(viewDimension)
			if err != nil {
				channels.ReportError(err)
				break
			}

			if err := grv.ui.Update(wins); err != nil {
				channels.ReportError(err)
				break
			}
		case err := <-errorCh:
			log.Errorf("Error channel received error: %v", err)
			errors = append(errors, err)

		OuterLoop:
			for {
				select {
				case err := <-errorCh:
					errors = append(errors, err)
					log.Errorf("Error channel received error: %v", err)
				default:
					break OuterLoop
				}
			}

			lastErrorReceivedTime = time.Now()

			if !timerActive {
				timer.Reset(grvMaxDrawFrequency)
				timerActive = true
			}
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}

func (grv *GRV) runHandlerLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, inputKeyCh <-chan string, actionCh chan Action, errorCh chan<- error, eventCh <-chan Event) {
	defer waitGroup.Done()
	defer log.Info("Handler loop stopping")
	log.Info("Starting handler loop")

	for {
		select {
		case key := <-inputKeyCh:
			grv.inputBuffer.Append(key)

			for {
				viewHierarchy := grv.view.ActiveViewIDHierarchy()
				action, keystring := grv.inputBuffer.Process(viewHierarchy)

				if action.ActionType != ActionNone {
					if IsPromptAction(action.ActionType) {
						keys, enterFound := grv.inputBuffer.DiscardTo("<Enter>")
						if enterFound {
							keys = strings.TrimSuffix(keys, "<Enter>")
						}

						action.Args = append(action.Args, ActionPromptArgs{
							keys:       keys,
							terminated: enterFound,
						})

						if err := grv.view.HandleAction(action); err != nil {
							errorCh <- err
						}
					} else {
						actionCh <- action
					}
				} else if keystring != "" {
					log.Debugf("Dropping keystring: %v", keystring)
				} else {
					break
				}
			}
		case action := <-actionCh:
			switch action.ActionType {
			case ActionExit:
				grv.End()
			case ActionSuspend:
				grv.Suspend()
			case ActionRunCommand:
				if err := grv.runCommand(action); err != nil {
					errorCh <- err
				}
			case ActionSleep:
				if err := grv.sleep(action); err != nil {
					errorCh <- err
				}
			default:
				if err := grv.view.HandleAction(action); err != nil {
					errorCh <- err
				}
			}
		case event := <-eventCh:
			log.Infof("Received event: %v", event)
			for _, eventListener := range grv.eventListeners {
				if err := eventListener.HandleEvent(event); err != nil {
					errorCh <- err
				}
			}
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}

func (grv *GRV) runCommand(action Action) (err error) {
	if len(action.Args) == 0 {
		return fmt.Errorf("Expected argument of type ActionRunCommandArgs")
	}

	arg, ok := action.Args[0].(ActionRunCommandArgs)
	if !ok {
		return fmt.Errorf("Expected argument of type ActionRunCommandArgs but found type %T", action.Args[0])
	}

	var cmd *exec.Cmd

	if arg.noShell {
		cmd = exec.Command(arg.command, arg.args...)
	} else {
		cmd = exec.Command("/bin/sh", "-c", arg.command)
	}

	if arg.stdin != nil {
		cmd.Stdin = arg.stdin
	}

	if arg.stdout != nil {
		cmd.Stdout = arg.stdout
	}

	if arg.stderr != nil {
		cmd.Stderr = arg.stderr
	}

	cmd.Env, cmd.Dir = grv.repoData.GenerateGitCommandEnvironment()

	if arg.interactive {
		grv.ui.Suspend()
	}

	if arg.beforeStart != nil {
		arg.beforeStart(cmd)
	}

	cmdError := cmd.Start()

	if cmdError == nil {
		if arg.onStart != nil {
			arg.onStart(cmd)
		}

		cmdError = cmd.Wait()
	}

	if arg.interactive {
		if arg.promptForInput && grv.config.GetBool(CfInputPromptAfterCommand) {
			cmd.Stdout.Write([]byte("\nPress any key to continue"))
			bufio.NewReader(cmd.Stdin).ReadByte()
		}

		if err = grv.ui.Resume(); err != nil {
			return
		}
	}

	exitStatus := -1

	if cmdError != nil {
		if exitError, ok := cmdError.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			exitStatus = waitStatus.ExitStatus()
		}
	} else {
		waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitStatus = waitStatus.ExitStatus()
	}

	if arg.onComplete != nil {
		err = arg.onComplete(cmdError, exitStatus)
	}

	return
}

func (grv *GRV) sleep(action Action) (err error) {
	if len(action.Args) == 0 {
		return fmt.Errorf("Expected sleep seconds argument")
	}

	sleepSeconds, ok := action.Args[0].(float64)
	if !ok {
		return fmt.Errorf("Expected sleep seconds of type float64 but found type %T", action.Args[0])
	}

	log.Infof("Sleeping for %v seconds", sleepSeconds)
	time.Sleep(time.Duration(sleepSeconds*1000) * time.Millisecond)
	log.Infof("Finished sleeping")

	return
}

func (grv *GRV) runSignalHandlerLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool) {
	defer waitGroup.Done()
	defer log.Info("Signal handler loop stopping")
	log.Info("Signal handler loop starting")

	signalCh := make(chan os.Signal, 1)

	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGWINCH, syscall.SIGCONT)

	for {
		select {
		case signal := <-signalCh:
			log.Debugf("Caught signal: %v", signal)

			switch signal {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP:
				if signal == syscall.SIGINT && ReadLineActive() {
					log.Debugf("Readline is active - cancelling readline")
					CancelReadline()
				} else {
					grv.End()
					return
				}
			case syscall.SIGCONT:
				grv.Resume()
			case syscall.SIGWINCH:
				if err := grv.ui.Resize(); err != nil {
					log.Errorf("Unable to resize display: %v", err)
				}

				grv.channels.displayCh <- true
			}
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}

func (grv *GRV) runFileSystemMonitorLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool) {
	defer waitGroup.Done()
	defer log.Info("FileSystem Monitor loop stopping")
	log.Info("FileSystem loop starting")

	channels := grv.channels.Channels()
	eventCh := make(chan fs.EventInfo, 1)
	repoGitDir := grv.repoData.Path()
	repoFilePath := grv.repoData.RepositoryRootPath()
	watchDir := repoFilePath + "..."

	if err := fs.Watch(watchDir, eventCh, fs.All); err != nil {
		log.Errorf("Unable to watch path for filesystem events %v: %v", watchDir, err)
		return
	}

	defer fs.Stop(eventCh)

	log.Infof("Watching filesystem events for path: %v", watchDir)

	timer := time.NewTimer(time.Hour)
	timer.Stop()
	timerActive := false

	ignorePaths := map[string]bool{}

	logFile := LogFile()
	if logFile != "" {
		ignorePaths[logFile] = true
		log.Debugf("Ignoring filesystem events for log file %v", logFile)

		if canonicalLogFile, err := CanonicalPath(logFile); err == nil {
			ignorePaths[canonicalLogFile] = true
			log.Debugf("Ignoring filesystem events for log file canonical path %v", canonicalLogFile)
		}
	}

	gitDirModified := false

	for {
		select {
		case event := <-eventCh:
			if _, ignore := ignorePaths[event.Path()]; !ignore {
				log.Debugf("FileSystem event: %v", event)

				if !timerActive {
					timer.Reset(grvMaxGitStatusFrequency)
					timerActive = true
				}

				if !gitDirModified && strings.HasPrefix(event.Path(), repoGitDir) {
					gitDirModified = true
				}
			}
		case <-timer.C:
			timerActive = false

			if gitDirModified {
				grv.repoData.Reload(nil)
				gitDirModified = false
			} else if err := grv.repoData.LoadStatus(); err != nil {
				channels.ReportError(err)
			}
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}
