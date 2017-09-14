package main

// #cgo darwin CFLAGS: -I/usr/local/opt/readline/include
// #cgo darwin LDFLAGS: -L/usr/local/opt/readline/lib
// #cgo LDFLAGS: -lreadline -lhistory
//
// // TODO: Find way of determining paths above in Makefile and providing them
// // as build flags
//
//
// #include <stdio.h>
// #include <stdlib.h>
// #include <readline/readline.h>
// #include <readline/history.h>
//
// extern void grvReadlineUpdateDisplay(void);
//
// static void grv_init_readline(void) {
// 	rl_redisplay_function = grvReadlineUpdateDisplay;
//	rl_catch_signals = 0;
//	rl_catch_sigwinch = 0;
//	rl_change_environment = 0;
//	rl_bind_key('\t', NULL);
//
//	history_write_timestamps = 1;
//	history_comment_char = '#';
//	using_history();
// }
import "C"

import (
	"os"
	"sync"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

const (
	rlCommandHistoryFile = "/command_history"
	rlSearchHistoryFile  = "/search_history"
	rlFilterHistoryFile  = "/filter_history"
)

var historyFilePrompts = map[string]string{
	PromptText:              rlCommandHistoryFile,
	SearchPromptText:        rlSearchHistoryFile,
	ReverseSearchPromptText: rlSearchHistoryFile,
	FilterPromptText:        rlFilterHistoryFile,
}

var readLine ReadLine

// ReadLine is a wrapper around the readline library
type ReadLine struct {
	channels       *Channels
	ui             InputUI
	config         Config
	promptText     string
	promptInput    string
	promptPoint    int
	active         bool
	lastPromptText string
	lock           sync.Mutex
}

// InitReadLine initialises the readline library
func InitReadLine(channels *Channels, ui InputUI, config Config) {
	readLine = ReadLine{
		channels: channels,
		config:   config,
		ui:       ui,
	}

	C.grv_init_readline()
}

// FreeReadLine flushes any history to disk
func FreeReadLine() {
	historyFile, hasHistoryFile := historyFilePrompts[readLine.lastPromptText]

	if hasHistoryFile {
		writeHistoryFile(historyFile)
	}
}

func readHistoryFile(file string) {
	configDir := readLine.config.ConfigDir()
	if configDir == "" {
		return
	}

	historyFilePath := configDir + file
	if _, err := os.Stat(historyFilePath); os.IsNotExist(err) {
		return
	}

	cHistoryFilePath := C.CString(historyFilePath)

	if C.read_history(cHistoryFilePath) != 0 {
		log.Errorf("Failed to load command history file %v", cHistoryFilePath)
	}

	C.free(unsafe.Pointer(cHistoryFilePath))
}

func writeHistoryFile(file string) {
	configDir := readLine.config.ConfigDir()
	if configDir == "" {
		return
	}

	cHistoryFilePath := C.CString(configDir + file)

	if C.write_history(cHistoryFilePath) != 0 {
		log.Errorf("Failed to write command history to file %v", cHistoryFilePath)
	}

	C.free(unsafe.Pointer(cHistoryFilePath))
}

// Prompt shows a readline prompt using prompt text provided
// User input is returned
func Prompt(prompt string) string {
	cPrompt := C.CString(prompt)

	readLineSetupPromptHistory(prompt)
	readLineSetActive(true)
	cInput := C.readline(cPrompt)
	readLineSetActive(false)

	C.free(unsafe.Pointer(cPrompt))
	readLineAddPromptHistory(prompt, cInput)
	input := C.GoString(cInput)
	C.free(unsafe.Pointer(cInput))

	return input
}

// PromptState returns current prompt properties
func PromptState() (string, string, int) {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	return readLine.promptText, readLine.promptInput, readLine.promptPoint
}

// ReadLineActive returns true if the readline prompt is currently displayed
func ReadLineActive() bool {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	return readLine.active
}

func readLineSetActive(active bool) {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	readLine.active = active
}

func readLineSetupPromptHistory(prompt string) {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	if prompt == readLine.lastPromptText {
		return
	}

	prevHistoryFile, ok := historyFilePrompts[readLine.lastPromptText]
	if ok {
		writeHistoryFile(prevHistoryFile)
	}

	C.clear_history()

	historyFile, ok := historyFilePrompts[prompt]
	if ok {
		readHistoryFile(historyFile)
	}
}

func readLineAddPromptHistory(prompt string, cInput *C.char) {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	readLine.lastPromptText = prompt

	if C.GoString(cInput) == "" {
		return
	}

	_, hasHistoryFile := historyFilePrompts[readLine.lastPromptText]

	if hasHistoryFile {
		C.add_history(cInput)
	}
}

//export grvReadlineUpdateDisplay
func grvReadlineUpdateDisplay() {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	displayPrompt := C.GoString(C.rl_display_prompt)
	lineBuffer := C.GoString(C.rl_line_buffer)
	point := int(C.rl_point)

	readLine.promptText = displayPrompt
	readLine.promptInput = lineBuffer
	readLine.promptPoint = point

	log.Debugf("ReadLine update display - prompt: %v%v, point: %v",
		readLine.promptText, readLine.promptInput, readLine.promptPoint)

	readLine.channels.UpdateDisplay()
}
