package main

// #cgo LDFLAGS: -lreadline
//
// #include <stdio.h>
// #include <stdlib.h>
// #include <readline/readline.h>
//
// extern void grvReadLineUpdateDisplay(void);
// extern int grvReadLineGetc(FILE *file);
//
// static void grvInitReadLine() {
// 	rl_redisplay_function = grvReadLineUpdateDisplay;
// 	rl_getc_function = grvReadLineGetc;
// }
import "C"

import (
	log "github.com/Sirupsen/logrus"
	"sync"
	"unsafe"
)

var readLine ReadLine

type BlockingInputInterceptor struct {
	inputCh chan Key
	lock    sync.Mutex
}

func NewBlockingInputInterceptor() *BlockingInputInterceptor {
	return &BlockingInputInterceptor{
		inputCh: make(chan Key),
	}
}

func (interceptor *BlockingInputInterceptor) GetInput() Key {
	return <-interceptor.inputCh
}

func (interceptor *BlockingInputInterceptor) InputConsumed(key Key) bool {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()

	if interceptor.inputCh == nil {
		return false
	}

	select {
	case interceptor.inputCh <- key:
	default:
	}

	return true
}

func (interceptor *BlockingInputInterceptor) Close() {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()

	close(interceptor.inputCh)
	interceptor.inputCh = nil
}

type ReadLine struct {
	channels    *Channels
	ui          InputUI
	interceptor *BlockingInputInterceptor
	promptText  string
	promptPoint int
}

func InitReadLine(channels *Channels, ui InputUI) {
	readLine = ReadLine{
		channels: channels,
		ui:       ui,
	}

	C.grvInitReadLine()
}

//export grvReadLineUpdateDisplay
func grvReadLineUpdateDisplay() {
	displayPrompt := C.GoString(C.rl_display_prompt)
	lineBuffer := C.GoString(C.rl_line_buffer)
	point := int(C.rl_point)

	readLine.promptText = displayPrompt + lineBuffer
	readLine.promptPoint = point

	log.Debugf("ReadLine update display - prompt: %v, point: %v",
		readLine.promptText, readLine.promptPoint)

	readLine.channels.UpdateDisplay()
}

//export grvReadLineGetc
func grvReadLineGetc(file *C.FILE) C.int {
	char := C.int(readLine.interceptor.GetInput())
	log.Debugf("ReadLine getc: %v", char)
	return char
}

func Prompt(prompt string) string {
	readLine.interceptor = NewBlockingInputInterceptor()
	readLine.ui.RegisterInputInterceptor(readLine.interceptor)

	cPrompt := C.CString(prompt)
	cInput := C.readline(cPrompt)
	C.free(unsafe.Pointer(cPrompt))

	input := C.GoString(cInput)
	C.free(unsafe.Pointer(cInput))

	readLine.ui.DeRegisterInputInterceptor(readLine.interceptor)
	readLine.interceptor.Close()
	readLine.interceptor = nil

	return input
}

func PromptState() (string, int) {
	return readLine.promptText, readLine.promptPoint
}
