package main

import (
	log "github.com/Sirupsen/logrus"
	"sync"
)

const (
	GRV_INPUT_BUFFER_SIZE   = 100
	GRV_DISPLAY_BUFFER_SIZE = 10
	GRV_ERROR_BUFFER_SIZE   = 10
)

type GRVChannels struct {
	exitCh    chan bool
	inputCh   chan KeyPressEvent
	displayCh chan bool
	errorCh   chan error
}

type HandlerChannels struct {
	displayCh chan<- bool
}

type GRV struct {
	repoData *RepositoryData
	view     *View
	ui       UI
	channels GRVChannels
}

func NewGRV() *GRV {
	repoDataLoader := NewRepoDataLoader()
	repoData := NewRepositoryData(repoDataLoader)

	return &GRV{
		repoData: repoData,
		view:     NewView(repoData),
		ui:       NewNcursesDisplay(),
		channels: GRVChannels{
			exitCh:    make(chan bool),
			inputCh:   make(chan KeyPressEvent, GRV_INPUT_BUFFER_SIZE),
			displayCh: make(chan bool, GRV_DISPLAY_BUFFER_SIZE),
			errorCh:   make(chan error, GRV_ERROR_BUFFER_SIZE),
		},
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

	if err = grv.view.Initialise(HandlerChannels{displayCh: grv.channels.displayCh}); err != nil {
		return
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
			inputCh <- keyPressEvent
			log.Debugf("Received keypress from UI %v", keyPressEvent)
		}
	}
}

func (grv *GRV) runDisplayLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, displayCh <-chan bool, errorCh chan error) {
	defer waitGroup.Done()
	defer log.Info("Display loop stopping")
	log.Info("Starting display loop")

	for {
		select {
		case <-displayCh:
			log.Debug("Received display refresh request")

			viewDimension := grv.ui.ViewDimension()

			wins, err := grv.view.Render(viewDimension)
			if err != nil {
				select {
				case errorCh <- err:
				default:
					log.Errorf("Unable to send error %v", err)
				}

				break
			}

			if err := grv.ui.Update(wins); err != nil {
				select {
				case errorCh <- err:
				default:
					log.Errorf("Unable to send error %v", err)
				}

				break
			}
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

	channels := HandlerChannels{
		displayCh: displayCh,
	}

	for {
		select {
		case keyPressEvent := <-inputCh:
			if err := grv.view.Handle(keyPressEvent, channels); err != nil {
				errorCh <- err
			}
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}
