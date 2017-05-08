package main

// #cgo LDFLAGS: -lreadline
//
// #include <stdio.h>
// #include <stdlib.h>
// #include <readline/readline.h>
//
// extern void grv_readline_update_display(void);
//
// static void grv_init_readline() {
// 	rl_redisplay_function = grv_readline_update_display;
//	rl_catch_signals = 0;
//	rl_catch_sigwinch = 0;
//      rl_change_environment = 0;
// }
import "C"

import (
	log "github.com/Sirupsen/logrus"
	"sync"
	"unsafe"
)

var readLine ReadLine

type ReadLine struct {
	channels    *Channels
	ui          InputUI
	promptText  string
	promptPoint int
	active      bool
	lock        sync.Mutex
}

func InitReadLine(channels *Channels, ui InputUI) {
	readLine = ReadLine{
		channels: channels,
		ui:       ui,
	}

	C.grv_init_readline()
}

func Prompt(prompt string) string {
	cPrompt := C.CString(prompt)

	readLineSetActive(true)
	cInput := C.readline(cPrompt)
	readLineSetActive(false)

	C.free(unsafe.Pointer(cPrompt))
	input := C.GoString(cInput)
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
