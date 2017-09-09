package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	grvInputBufferSize   = 100
	grvActionBufferSize  = 100
	grvErrorBufferSize   = 100
	grvDisplayBufferSize = 50
	grvMaxDrawFrequency  = time.Millisecond * 50
	grvMinErrorDisplay   = time.Second * 2
)

type gRVChannels struct {
	exitCh     chan bool
	inputKeyCh chan string
	actionCh   chan Action
	displayCh  chan bool
	errorCh    chan error
}

func (grvChannels gRVChannels) Channels() *Channels {
	return &Channels{
		displayCh: grvChannels.displayCh,
		exitCh:    grvChannels.exitCh,
		errorCh:   grvChannels.errorCh,
		actionCh:  grvChannels.actionCh,
	}
}

// Channels contains channels used for communication within grv
type Channels struct {
	displayCh chan<- bool
	exitCh    <-chan bool
	errorCh   chan<- error
	actionCh  chan<- Action
}

// GRV is the top level structure containing all state in the program
type GRV struct {
	repoData    *RepositoryData
	view        *View
	ui          UI
	channels    gRVChannels
	config      *Configuration
	inputBuffer *InputBuffer
	input       *InputKeyMapper
}

// UpdateDisplay sends a request to update the display
func (channels *Channels) UpdateDisplay() {
	select {
	case channels.displayCh <- true:
	default:
	}
}

// Exit returns true if GRV is in the process of exiting
func (channels *Channels) Exit() bool {
	select {
	case _, ok := <-channels.exitCh:
		return !ok
	default:
		return false
	}
}

// ReportError reports an error to be displayed
func (channels *Channels) ReportError(err error) {
	if err != nil {
		select {
		case channels.errorCh <- err:
		default:
			log.Errorf("Unable to report error %v", err)
		}
	}
}

// ReportErrors reports multiple errors to be displayed
func (channels *Channels) ReportErrors(errors []error) {
	if errors == nil {
		return
	}

	for _, err := range errors {
		channels.ReportError(err)
	}
}

// DoAction sends an action to be executed
func (channels *Channels) DoAction(action Action) {
	if action.ActionType != ActionNone {
		channels.actionCh <- action
	}
}

// ReportStatus updates the status bar with the provided status
func (channels *Channels) ReportStatus(format string, args ...interface{}) {
	status := fmt.Sprintf(format, args...)

	if status != "" {
		channels.DoAction(Action{
			ActionType: ActionShowStatus,
			Args:       []interface{}{status},
		})
	}
}

// NewGRV creates a new instace of GRV
func NewGRV() *GRV {
	grvChannels := gRVChannels{
		exitCh:     make(chan bool),
		inputKeyCh: make(chan string, grvInputBufferSize),
		actionCh:   make(chan Action, grvActionBufferSize),
		displayCh:  make(chan bool, grvDisplayBufferSize),
		errorCh:    make(chan error, grvErrorBufferSize),
	}

	channels := grvChannels.Channels()

	repoDataLoader := NewRepoDataLoader(channels)
	repoData := NewRepositoryData(repoDataLoader, channels)
	keyBindings := NewKeyBindingManager()
	config := NewConfiguration(keyBindings, channels)
	ui := NewNCursesDisplay(config)

	return &GRV{
		repoData:    repoData,
		view:        NewView(repoData, channels, config),
		ui:          ui,
		channels:    grvChannels,
		config:      config,
		inputBuffer: NewInputBuffer(keyBindings),
		input:       NewInputKeyMapper(ui),
	}
}

// Initialise sets up all the components of GRV
func (grv *GRV) Initialise(repoPath string) (err error) {
	log.Info("Initialising GRV")

	if err = grv.repoData.Initialise(repoPath); err != nil {
		return
	}

	if err = grv.ui.Initialise(); err != nil {
		return
	}

	if err = grv.view.Initialise(); err != nil {
		return
	}

	if configErrors := grv.config.Initialise(); configErrors != nil {
		for _, configError := range configErrors {
			grv.channels.errorCh <- configError
		}
	}

	channels := grv.channels.Channels()
	InitReadLine(channels, grv.ui, grv.config)

	return
}

// Free closes and frees any resources used by GRV
func (grv *GRV) Free() {
	log.Info("Freeing GRV")

	FreeReadLine()
	grv.ui.Free()
	grv.repoData.Free()
}

// Suspend prepares GRV to be suspended and sends a SIGTSTP
// to every process in the process group
func (grv *GRV) Suspend() {
	log.Info("Suspending GRV")

	grv.ui.Suspend()
	if err := syscall.Kill(0, syscall.SIGTSTP); err != nil {
		log.Errorf("Error when attempting to suspend GRV: %v", err)
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
	go grv.runHandlerLoop(&waitGroup, channels.exitCh, channels.inputKeyCh, channels.actionCh, channels.errorCh)
	waitGroup.Add(1)
	go grv.runSignalHandlerLoop(&waitGroup, channels.exitCh)

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

	displayTimerCh := time.NewTicker(grvMaxDrawFrequency)
	defer displayTimerCh.Stop()
	refreshRequestReceived := false
	channels := &Channels{errorCh: errorCh}

	var errors []error
	lastErrorReceivedTime := time.Now()

	for {
		select {
		case <-displayCh:
			log.Debug("Received display refresh request")
			refreshRequestReceived = true
		case <-displayTimerCh.C:
			if !refreshRequestReceived {
				break
			}

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

			refreshRequestReceived = false
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
			refreshRequestReceived = true
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}

func (grv *GRV) runHandlerLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, inputKeyCh <-chan string, actionCh chan Action, errorCh chan<- error) {
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
					actionCh <- action
				} else if keystring != "" {
					if err := grv.view.HandleKeyPress(keystring); err != nil {
						errorCh <- err
					}
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
			default:
				if err := grv.view.HandleAction(action); err != nil {
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
				grv.End()
				return
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
