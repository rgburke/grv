package main

import (
	"sync"

	log "github.com/Sirupsen/logrus"
)

// SearchableView is a view that supports searching functionality
type SearchableView interface {
	SearchInputProvidor
	ViewPos() ViewPos
	OnSearchMatch(startPos ViewPos, matchLineIndex uint)
}

// ViewSearch manages search functionality for a view
type ViewSearch struct {
	search               *Search
	searchableView       SearchableView
	channels             Channels
	lastSearchFoundMatch bool
	lock                 sync.Mutex
}

// NewViewSearch creates a new instance
func NewViewSearch(searchableView SearchableView, channels Channels) *ViewSearch {
	return &ViewSearch{
		searchableView: searchableView,
		channels:       channels,
	}
}

// SearchActive returns the state of the most recent search (if one has been performed)
func (viewSearch *ViewSearch) SearchActive() (active bool, pattern string, lastSearchFoundMatch bool) {
	viewSearch.lock.Lock()
	defer viewSearch.lock.Unlock()

	return viewSearch.searchActive()
}

func (viewSearch *ViewSearch) searchActive() (active bool, pattern string, lastSearchFoundMatch bool) {
	if viewSearch.search != nil {
		active = true
		pattern = viewSearch.search.pattern
		lastSearchFoundMatch = viewSearch.lastSearchFoundMatch
	}

	return
}

// HandleAction handles all actions relating to search that a view receives
func (viewSearch *ViewSearch) HandleAction(action Action) (handled bool, err error) {
	viewSearch.lock.Lock()
	defer viewSearch.lock.Unlock()

	handled = true

	switch action.ActionType {
	case ActionSearch, ActionReverseSearch:
		err = viewSearch.doSearch(action)
	case ActionSearchFindNext:
		err = viewSearch.findNextMatch()
	case ActionSearchFindPrev:
		err = viewSearch.findPrevMatch()
	case ActionClearSearch:
		err = viewSearch.clearSearch()
	default:
		handled = false
	}

	return
}

func (viewSearch *ViewSearch) doSearch(action Action) (err error) {
	search, err := CreateSearchFromAction(action, viewSearch.searchableView)
	if err != nil {
		return
	}

	viewSearch.search = search

	return viewSearch.findNextMatch()
}

func (viewSearch *ViewSearch) findNextMatch() (err error) {
	active, pattern, _ := viewSearch.searchActive()
	if !active {
		return
	}

	viewSearch.channels.ReportStatus("Searching...")

	go func() {
		viewPos := viewSearch.searchableView.ViewPos()
		log.Debugf("Searching for next occurrence of pattern %v starting from row index :%v",
			pattern, viewPos.ActiveRowIndex())

		matchLineIndex, found := viewSearch.search.FindNext(viewPos.ActiveRowIndex())

		viewSearch.lock.Lock()
		viewSearch.lastSearchFoundMatch = found
		viewSearch.lock.Unlock()

		if found {
			viewSearch.searchableView.OnSearchMatch(viewPos, matchLineIndex)
			viewSearch.channels.ReportStatus("Match found")
		} else {
			viewSearch.channels.ReportStatus("No matches found")
		}
	}()

	return
}

func (viewSearch *ViewSearch) findPrevMatch() (err error) {
	active, pattern, _ := viewSearch.searchActive()
	if !active {
		return
	}

	viewSearch.channels.ReportStatus("Searching...")

	go func() {
		viewPos := viewSearch.searchableView.ViewPos()
		log.Debugf("Searching for previous occurrence of pattern %v starting from row index :%v",
			pattern, viewPos.ActiveRowIndex())

		matchLineIndex, found := viewSearch.search.FindPrev(viewPos.ActiveRowIndex())

		viewSearch.lock.Lock()
		viewSearch.lastSearchFoundMatch = found
		viewSearch.lock.Unlock()

		if found {
			viewSearch.searchableView.OnSearchMatch(viewPos, matchLineIndex)
			viewSearch.channels.ReportStatus("Match found")
		} else {
			viewSearch.channels.ReportStatus("No matches found")
		}

	}()

	return
}

func (viewSearch *ViewSearch) clearSearch() (err error) {
	if active, pattern, _ := viewSearch.searchActive(); active {
		viewSearch.channels.ReportStatus("Cleared search")
		log.Debugf("Clearing search with pattern %v", pattern)
		viewSearch.search = nil
		viewSearch.lastSearchFoundMatch = false
	}

	return
}
