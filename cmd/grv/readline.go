package main

// #cgo darwin CFLAGS: -I/usr/local/opt/readline/include
// #cgo freebsd CFLAGS: -I/usr/local/include
// #cgo darwin LDFLAGS: -L/usr/local/opt/readline/lib
// #cgo freebsd LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lreadline -lhistory
//
// #include <stdio.h>
// #include <stdlib.h>
// #include <readline/readline.h>
// #include <readline/history.h>
//
// extern void grvReadlineUpdateDisplay(void);
// extern int grvReadlineStartUpHook(void);
//
// static void grv_init_readline(void) {
// 	rl_redisplay_function = grvReadlineUpdateDisplay;
//	rl_startup_hook = grvReadlineStartUpHook;
//	rl_catch_signals = 0;
//	rl_catch_sigwinch = 0;
//#if RL_READLINE_VERSION >= 0x0603
//	rl_change_environment = 0;
//#endif
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
	rlCommandHistoryFile    = "/command_history"
	rlSearchHistoryFile     = "/search_history"
	rlFilterHistoryFile     = "/filter_history"
	rlBranchNameHistoryFile = "/branch_name_history"
)

var historyFilePrompts = map[string]string{
	PromptText:              rlCommandHistoryFile,
	SearchPromptText:        rlSearchHistoryFile,
	ReverseSearchPromptText: rlSearchHistoryFile,
	FilterPromptText:        rlFilterHistoryFile,
	BranchNamePromptText:    rlBranchNameHistoryFile,
}

// PromptArgs contains arguments to configure the display of a prompt
type PromptArgs struct {
	Prompt            string
	InitialBufferText string
	NumCharsToRead    int
}

var readLine ReadLine

// ReadLine is a wrapper around the readline library
type ReadLine struct {
	channels          *Channels
	config            Config
	promptText        string
	promptInput       string
	promptPoint       int
	active            bool
	lastPromptText    string
	initialBufferText string
	lock              sync.Mutex
}

// InitReadLine initialises the readline library
func InitReadLine(channels *Channels, config Config) {
	readLine = ReadLine{
		channels: channels,
		config:   config,
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

// Prompt shows a readline prompt using the args provided
// User input is returned
func Prompt(promptArgs *PromptArgs) string {
	if promptArgs.InitialBufferText != "" {
		readLineSetInitialBufferText(promptArgs.InitialBufferText)
		defer readLineSetInitialBufferText("")
	}

	if promptArgs.NumCharsToRead > 0 {
		readLineSetNumCharsToRead(promptArgs.NumCharsToRead)
		defer readLineSetNumCharsToRead(0)
	}

	readLineSetupPromptHistory(promptArgs.Prompt)
	readLineSetActive(true)
	cPrompt := C.CString(promptArgs.Prompt)
	cInput := C.readline(cPrompt)
	readLineSetActive(false)

	C.free(unsafe.Pointer(cPrompt))
	readLineAddPromptHistory(promptArgs.Prompt, cInput)
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

func readLineSetInitialBufferText(initialBufferText string) {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	readLine.initialBufferText = initialBufferText
}

func readLineSetNumCharsToRead(numCharsToRead int) {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	C.rl_num_chars_to_read = C.int(numCharsToRead)
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

//export grvReadlineStartUpHook
func grvReadlineStartUpHook() C.int {
	readLine.lock.Lock()
	defer readLine.lock.Unlock()

	if readLine.initialBufferText != "" {
		cInitialBufferText := C.CString(readLine.initialBufferText)
		C.rl_insert_text(cInitialBufferText)
		C.free(unsafe.Pointer(cInitialBufferText))
	}

	return 0
}
