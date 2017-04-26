package main

import (
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
)

const (
	GRV_INPUT_BUFFER_SIZE  = 100
	GRV_ERROR_BUFFER_SIZE  = 100
	GRV_MAX_DRAW_FREQUENCY = time.Millisecond * 50
)

type GRVChannels struct {
	exitCh     chan bool
	inputKeyCh chan string
	displayCh  chan bool
	errorCh    chan error
}

type Channels struct {
	displayCh chan<- bool
	exitCh    <-chan bool
	errorCh   chan<- error
}

type GRV struct {
	repoData    *RepositoryData
	view        *View
	ui          UI
	channels    GRVChannels
	config      *Configuration
	inputBuffer *InputBuffer
	input       *InputKeyMapper
}

func (channels *Channels) UpdateDisplay() {
	select {
	case channels.displayCh <- true:
	default:
	}
}

// Check if grv is exiting
// This is intended to be used by long running go routines
func (channels *Channels) Exit() bool {
	select {
	case _, ok := <-channels.exitCh:
		return !ok
	default:
		return false
	}
}

// Report an error to the error channel
// This is intended to be used by go routines to report errors that cannot be returned
func (channels *Channels) ReportError(err error) {
	if err != nil {
		select {
		case channels.errorCh <- err:
		default:
			log.Errorf("Unable to report error %v", err)
		}
	}
}

func NewGRV() *GRV {
	grvChannels := GRVChannels{
		exitCh:     make(chan bool),
		inputKeyCh: make(chan string, GRV_INPUT_BUFFER_SIZE),
		displayCh:  make(chan bool),
		errorCh:    make(chan error, GRV_ERROR_BUFFER_SIZE),
	}

	channels := &Channels{
		displayCh: grvChannels.displayCh,
		exitCh:    grvChannels.exitCh,
		errorCh:   grvChannels.errorCh,
	}

	repoDataLoader := NewRepoDataLoader(channels)
	repoData := NewRepositoryData(repoDataLoader, channels)
	keyBindings := NewKeyBindingManager()
	config := NewConfiguration(keyBindings)
	ui := NewNcursesDisplay(config)

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

	return
}

func (grv *GRV) Free() {
	log.Info("Freeing GRV")

	grv.ui.Free()
	grv.repoData.Free()
}

func (grv *GRV) Run() {
	var waitGroup sync.WaitGroup
	channels := grv.channels

	waitGroup.Add(1)
	go grv.runInputLoop(&waitGroup, channels.exitCh, channels.inputKeyCh, channels.errorCh)
	waitGroup.Add(1)
	go grv.runDisplayLoop(&waitGroup, channels.exitCh, channels.displayCh, channels.errorCh)
	waitGroup.Add(1)
	go grv.runHandlerLoop(&waitGroup, channels.exitCh, channels.displayCh, channels.inputKeyCh, channels.errorCh)

	channels.displayCh <- true

	log.Info("Waiting for loops to finish")
	waitGroup.Wait()
	log.Info("All loops finished")
}

func (grv *GRV) runInputLoop(waitGroup *sync.WaitGroup, exitCh chan<- bool, inputKeyCh chan<- string, errorCh chan<- error) {
	defer waitGroup.Done()
	defer log.Info("Input loop stopping")
	log.Info("Starting input loop")

	for {
		key, err := grv.input.GetKeyInput()
		if err != nil {
			errorCh <- err
		} else if key == "q" {
			log.Infof("Received exit key %v, now closing exit channel", key)
			close(exitCh)
			return
		} else {
			log.Debugf("Received keypress from UI %v", key)

			select {
			case inputKeyCh <- key:
			default:
				log.Errorf("Unable to add keypress %v to input channel", key)
			}
		}
	}
}

func (grv *GRV) runDisplayLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, displayCh <-chan bool, errorCh chan error) {
	defer waitGroup.Done()
	defer log.Info("Display loop stopping")
	log.Info("Starting display loop")

	displayTimerCh := time.NewTicker(GRV_MAX_DRAW_FREQUENCY)
	defer displayTimerCh.Stop()
	refreshRequestReceived := false
	channels := &Channels{errorCh: errorCh}

	for {
		select {
		case <-displayCh:
			log.Debug("Received display refresh request")
			refreshRequestReceived = true
		case <-displayTimerCh.C:
			if !refreshRequestReceived {
				break
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
			grv.ui.ShowError(err)
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}

func (grv *GRV) runHandlerLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, displayCh chan<- bool, inputKeyCh <-chan string, errorCh chan<- error) {
	defer waitGroup.Done()
	defer log.Info("Handler loop stopping")
	log.Info("Starting handler loop")

	for {
		select {
		case key := <-inputKeyCh:
			grv.inputBuffer.Append(key)

			for {
				viewHierarchy := grv.view.ActiveViewHierarchy()
				action, keystring := grv.inputBuffer.Process(viewHierarchy)

				if action != ACTION_NONE {
					if err := grv.view.HandleAction(action); err != nil {
						errorCh <- err
					}
				} else if keystring != "" {
					if err := grv.view.HandleKeyPress(keystring); err != nil {
						errorCh <- err
					}
				} else {
					break
				}
			}
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}
