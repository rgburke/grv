package main

import (
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
)

const (
	INPUT_BUFFER_SIZE   = 100
	DISPLAY_BUFFER_SIZE = 10
	ERROR_BUFFER_SIZE   = 10
	INPUT_SLEEP_MS      = 100
)

type HandlerChannels struct {
	displayCh chan<- bool
	inputCh   <-chan KeyPressEvent
}

type GRV struct {
	repoData *RepositoryData
	view     *View
	ui       UI
}

func NewGRV() *GRV {
	repoDataLoader := NewRepoDataLoader()
	repoData := NewRepositoryData(repoDataLoader)

	return &GRV{
		repoData: repoData,
		view:     NewView(repoData),
		ui:       NewNcursesDisplay(),
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

	return
}

func (grv *GRV) Free() {
	log.Info("Freeing GRV")

	grv.ui.Free()
	grv.repoData.Free()
}

func (grv *GRV) Run() {
	exitCh := make(chan bool)
	inputCh := make(chan KeyPressEvent, INPUT_BUFFER_SIZE)
	displayCh := make(chan bool, DISPLAY_BUFFER_SIZE)
	errorCh := make(chan error, ERROR_BUFFER_SIZE)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go grv.runInputLoop(&waitGroup, exitCh, inputCh, errorCh)
	waitGroup.Add(1)
	go grv.runDisplayLoop(&waitGroup, exitCh, displayCh, errorCh)
	waitGroup.Add(1)
	go grv.runHandlerLoop(&waitGroup, exitCh, displayCh, inputCh, errorCh)

	displayCh <- true

	log.Info("Waiting for loops to finish")
	waitGroup.Wait()
	log.Info("All loops finished")
}

func (grv *GRV) runInputLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, inputCh chan<- KeyPressEvent, errorCh chan<- error) {
	defer waitGroup.Done()
	defer log.Info("Input loop stopping")
	log.Info("Starting input loop")

	for {
		keyPressEvent, err := grv.ui.GetInput()
		if err != nil {
			errorCh <- err
		} else if int(keyPressEvent.key) != 0 {
			inputCh <- keyPressEvent
			log.Debugf("Received keypress from UI %v", keyPressEvent)
		}

		select {
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		default:
			time.Sleep(INPUT_SLEEP_MS * time.Millisecond)
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
				errorCh <- err
			}

			if err := grv.ui.Update(wins); err != nil {
				errorCh <- err
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

func (grv *GRV) runHandlerLoop(waitGroup *sync.WaitGroup, exitCh chan<- bool, displayCh chan<- bool, inputCh <-chan KeyPressEvent, errorCh chan<- error) {
	defer waitGroup.Done()
	defer log.Info("Handler loop stopping")
	log.Info("Starting handler loop")

	channels := HandlerChannels{
		displayCh: displayCh,
		inputCh:   inputCh,
	}

	for {
		select {
		case keyPressEvent := <-inputCh:
			if keyPressEvent.key == 'q' {
				log.Infof("Received exit key %v, now closing exit channel", keyPressEvent)
				close(exitCh)
				return
			}

			if err := grv.view.Handle(keyPressEvent, channels); err != nil {
				errorCh <- err
			}
		}
	}
}
