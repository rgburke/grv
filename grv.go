package main

import (
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
)

const (
	GRV_INPUT_BUFFER_SIZE   = 100
	GRV_DISPLAY_BUFFER_SIZE = 10
	GRV_ERROR_BUFFER_SIZE   = 100
	GRV_MAX_DRAW_FREQUENCY  = time.Millisecond * 100
)

type GRVChannels struct {
	exitCh    chan bool
	inputCh   chan KeyPressEvent
	displayCh chan bool
	errorCh   chan error
}

type Channels struct {
	displayCh chan<- bool
	exitCh    <-chan bool
	errorCh   chan<- error
}

type GRV struct {
	repoData *RepositoryData
	view     *View
	ui       UI
	channels GRVChannels
	config   *Configuration
}

func (channels *Channels) UpdateDisplay() {
	channels.displayCh <- true
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
		exitCh:    make(chan bool),
		inputCh:   make(chan KeyPressEvent, GRV_INPUT_BUFFER_SIZE),
		displayCh: make(chan bool, GRV_DISPLAY_BUFFER_SIZE),
		errorCh:   make(chan error, GRV_ERROR_BUFFER_SIZE),
	}

	channels := &Channels{
		displayCh: grvChannels.displayCh,
		exitCh:    grvChannels.exitCh,
		errorCh:   grvChannels.errorCh,
	}

	repoDataLoader := NewRepoDataLoader(channels)
	repoData := NewRepositoryData(repoDataLoader, channels)
	config := NewConfiguration()

	return &GRV{
		repoData: repoData,
		view:     NewView(repoData, channels, config),
		ui:       NewNcursesDisplay(config),
		channels: grvChannels,
		config:   config,
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
	go grv.runInputLoop(&waitGroup, channels.exitCh, channels.inputCh, channels.errorCh)
	waitGroup.Add(1)
	go grv.runDisplayLoop(&waitGroup, channels.exitCh, channels.displayCh, channels.errorCh)
	waitGroup.Add(1)
	go grv.runHandlerLoop(&waitGroup, channels.exitCh, channels.displayCh, channels.inputCh, channels.errorCh)

	channels.displayCh <- true

	log.Info("Waiting for loops to finish")
	waitGroup.Wait()
	log.Info("All loops finished")
}

func (grv *GRV) runInputLoop(waitGroup *sync.WaitGroup, exitCh chan<- bool, inputCh chan<- KeyPressEvent, errorCh chan<- error) {
	defer waitGroup.Done()
	defer log.Info("Input loop stopping")
	log.Info("Starting input loop")

	for {
		keyPressEvent, err := grv.ui.GetInput()
		if err != nil {
			errorCh <- err
		} else if keyPressEvent.key == 'q' {
			log.Infof("Received exit key %v, now closing exit channel", keyPressEvent)
			close(exitCh)
			return
		} else if int(keyPressEvent.key) != 0 {
			log.Debugf("Received keypress from UI %v", keyPressEvent)

			select {
			case inputCh <- keyPressEvent:
			default:
				log.Errorf("Unable to add keypress %v to input channel", keyPressEvent)
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

func (grv *GRV) runHandlerLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, displayCh chan<- bool, inputCh <-chan KeyPressEvent, errorCh chan<- error) {
	defer waitGroup.Done()
	defer log.Info("Handler loop stopping")
	log.Info("Starting handler loop")

	for {
		select {
		case keyPressEvent := <-inputCh:
			if err := grv.view.Handle(keyPressEvent); err != nil {
				errorCh <- err
			}
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}
