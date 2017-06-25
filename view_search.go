package main

import (
	log "github.com/Sirupsen/logrus"
)

type SearchableView interface {
	SearchInputProvidor
	ViewPos() *ViewPos
	OnSearchMatch(startPos *ViewPos, matchLineIndex uint)
}

type ViewSearch struct {
	search         *Search
	searchableView SearchableView
	channels       *Channels
}

func NewViewSearch(searchableView SearchableView, channels *Channels) *ViewSearch {
	return &ViewSearch{
		searchableView: searchableView,
		channels:       channels,
	}
}

func (viewSearch *ViewSearch) HandleAction(action Action) (handled bool, err error) {
	handled = true

	switch action.ActionType {
	case ACTION_SEARCH, ACTION_REVERSE_SEARCH:
		err = viewSearch.doSearch(action)
	case ACTION_SEARCH_FIND_NEXT:
		err = viewSearch.findNextMatch()
	case ACTION_SEARCH_FIND_PREV:
		err = viewSearch.findPrevMatch()
	case ACTION_CLEAR_SEARCH:
		err = viewSearch.clearSearch()
	default:
		handled = false
	}

	return
}

func (viewSearch *ViewSearch) SearchActive() (active bool, pattern string) {
	if viewSearch.search != nil {
		active = true
		pattern = viewSearch.search.pattern
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
	active, pattern := viewSearch.SearchActive()
	if !active {
		return
	}

	viewPos := viewSearch.searchableView.ViewPos()

	viewSearch.channels.ReportStatus("Searching...")
	log.Debugf("Searching for next occurence of pattern %v starting from row index :%v",
		pattern, viewPos.activeRowIndex)

	go func() {
		matchLineIndex, found := viewSearch.search.FindNext(viewPos.activeRowIndex)

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
	active, pattern := viewSearch.SearchActive()
	if !active {
		return
	}

	viewPos := viewSearch.searchableView.ViewPos()

	viewSearch.channels.ReportStatus("Searching...")
	log.Debugf("Searching for previous occurence of pattern %v starting from row index :%v",
		pattern, viewPos.activeRowIndex)

	go func() {
		matchLineIndex, found := viewSearch.search.FindPrev(viewPos.activeRowIndex)

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
	if active, pattern := viewSearch.SearchActive(); active {
		viewSearch.channels.ReportStatus("Cleared search")
		log.Debugf("Clearing search with pattern %v", pattern)
		viewSearch.search = nil
	}

	return
}
