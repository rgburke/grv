package main

// #cgo LDFLAGS: -lreadline -lhistory
//
// #include <stdio.h>
// #include <stdlib.h>
// #include <readline/readline.h>
// #include <readline/history.h>
//
// extern void grv_readline_update_display(void);
//
// static void grv_init_readline(void) {
// 	rl_redisplay_function = grv_readline_update_display;
//	rl_catch_signals = 0;
//	rl_catch_sigwinch = 0;
//	rl_change_environment = 0;
//	using_history();
// }
import "C"

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"sync"
	"unsafe"
)

const (
	RL_COMMAND_HISTORY_FILE = "/history"
)

var readLine ReadLine

type ReadLine struct {
	channels    *Channels
	ui          InputUI
	config      Config
	promptText  string
	promptPoint int
	active      bool
	lock        sync.Mutex
}

func InitReadLine(channels *Channels, ui InputUI, config Config) {
	readLine = ReadLine{
		channels: channels,
		config:   config,
		ui:       ui,
	}

	C.grv_init_readline()
	readHistoryFile()
}

func FreeReadLine() {
	writeHistoryFile()
}

func readHistoryFile() {
	configDir := readLine.config.ConfigDir()
	if configDir == "" {
		return
	}

	historyFilePath := configDir + RL_COMMAND_HISTORY_FILE
	if _, err := os.Stat(historyFilePath); os.IsNotExist(err) {
		return
	}

	cHistoryFilePath := C.CString(historyFilePath)

	if C.read_history(cHistoryFilePath) != 0 {
		log.Errorf("Failed to load command history file %v", cHistoryFilePath)
	}

	C.free(unsafe.Pointer(cHistoryFilePath))
}

func writeHistoryFile() {
	configDir := readLine.config.ConfigDir()
	if configDir == "" {
		return
	}

	cHistoryFilePath := C.CString(configDir + RL_COMMAND_HISTORY_FILE)

	if C.write_history(cHistoryFilePath) != 0 {
		log.Errorf("Failed to write command history to file %v", cHistoryFilePath)
	}

	C.free(unsafe.Pointer(cHistoryFilePath))
}

func Prompt(prompt string) string {
	cPrompt := C.CString(prompt)

	readLineSetActive(true)
	cInput := C.readline(cPrompt)
	readLineSetActive(false)

	C.free(unsafe.Pointer(cPrompt))
	input := C.GoString(cInput)
	C.add_history(cInput)
	C.free(unsafe.Pointer(cInput))

	return input
}

func PromptState() (string, int) {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	return readLine.promptText, readLine.promptPoint
}

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

//export grv_readline_update_display
func grv_readline_update_display() {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	displayPrompt := C.GoString(C.rl_display_prompt)
	lineBuffer := C.GoString(C.rl_line_buffer)
	point := int(C.rl_point)

	readLine.promptText = displayPrompt + lineBuffer
	readLine.promptPoint = point

	log.Debugf("ReadLine update display - prompt: %v, point: %v",
		readLine.promptText, readLine.promptPoint)

	readLine.channels.UpdateDisplay()
}
