package main

import (
	"sync"
	"time"
)

const (
	INPUT_BUFFER_SIZE   = 100
	DISPLAY_BUFFER_SIZE = 10
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
	grv.ui.Free()
	grv.repoData.Free()
}

func (grv *GRV) Run() {
	exitCh := make(chan bool)
	inputCh := make(chan KeyPressEvent, INPUT_BUFFER_SIZE)
	displayCh := make(chan bool, DISPLAY_BUFFER_SIZE)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go grv.runInputLoop(&waitGroup, exitCh, inputCh)
	waitGroup.Add(1)
	go grv.runDisplayLoop(&waitGroup, exitCh, displayCh)
	waitGroup.Add(1)
	go grv.runHandlerLoop(&waitGroup, exitCh, displayCh, inputCh)

	displayCh <- true

	waitGroup.Wait()
}

func (grv *GRV) runInputLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, inputCh chan<- KeyPressEvent) {
	defer waitGroup.Done()

	for {
		keyPressEvent, err := grv.ui.GetInput()
		if err != nil {

		} else if int(keyPressEvent.key) != 0 {
			inputCh <- keyPressEvent
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

func (grv *GRV) runDisplayLoop(waitGroup *sync.WaitGroup, exitCh <-chan bool, displayCh <-chan bool) {
	defer waitGroup.Done()

	for {
		select {
		case <-displayCh:
			viewDimension := grv.ui.ViewDimension()

			wins, err := grv.view.Render(viewDimension)
			if err != nil {
				// TODO Handle
			}

			if err := grv.ui.Update(wins); err != nil {
				// TODO Handle
			}
		case _, ok := <-exitCh:
			if !ok {
				return
			}
		}
	}
}

func (grv *GRV) runHandlerLoop(waitGroup *sync.WaitGroup, exitCh chan<- bool, displayCh chan<- bool, inputCh <-chan KeyPressEvent) {
	defer waitGroup.Done()

	channels := HandlerChannels{
		displayCh: displayCh,
		inputCh:   inputCh,
	}

	for {
		select {
		case keyPressEvent := <-inputCh:
			if keyPressEvent.key == 'q' {
				close(exitCh)
				return
			}

			grv.view.Handle(keyPressEvent, channels)
		}
	}
}
