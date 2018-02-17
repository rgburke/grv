package main

import (
	"strings"
)

// InputBuffer buffers input and maps it to configured actions or key sequences
type InputBuffer struct {
	buffer      []string
	keyBindings KeyBindings
}

// NewInputBuffer creates a new input buffer instance
func NewInputBuffer(keyBindings KeyBindings) *InputBuffer {
	return &InputBuffer{
		buffer:      make([]string, 0),
		keyBindings: keyBindings,
	}
}

// Append adds new input to the end of the buffer
func (inputBuffer *InputBuffer) Append(input string) {
	keys := TokeniseKeys(input)
	inputBuffer.buffer = append(inputBuffer.buffer, keys...)
}

func (inputBuffer *InputBuffer) prepend(keys []string) {
	inputBuffer.buffer = append(keys, inputBuffer.buffer...)
}

func (inputBuffer *InputBuffer) pop() (key string) {
	key = inputBuffer.buffer[0]
	inputBuffer.buffer = inputBuffer.buffer[1:]
	return
}

func (inputBuffer *InputBuffer) hasInput() bool {
	return len(inputBuffer.buffer) > 0
}

// Process goes through the input in the buffer and attempts to map it to actions or key sequences
// If no mapping is possible the key sequences on the buffer are returned.
// If a prefix is matched then the buffer returns NOP so that more input can be appended to it
func (inputBuffer *InputBuffer) Process(viewHierarchy ViewHierarchy) (action Action, keystring string) {
	if !inputBuffer.hasInput() {
		return
	}

	keyBuffer := make([]string, 0)
	keyBindings := inputBuffer.keyBindings
	isPrefix := false

OuterLoop:
	for inputBuffer.hasInput() {
		keyBuffer = append(keyBuffer, inputBuffer.pop())
		binding, prefix := keyBindings.Binding(viewHierarchy, strings.Join(keyBuffer, ""))

		switch {
		case prefix:
			if len(inputBuffer.buffer) == 0 {
				inputBuffer.prepend(keyBuffer)
				return
			}

			isPrefix = true
		case binding.bindingType == BtAction:
			if binding.actionType != ActionNone {
				action = Action{ActionType: binding.actionType}
			} else if isPrefix {
				inputBuffer.prepend(keyBuffer[1:])
				keyBuffer = keyBuffer[0:1]
			}

			break OuterLoop
		case binding.bindingType == BtKeystring:
			inputBuffer.prepend(TokeniseKeys(binding.keystring))
			keyBuffer = keyBuffer[0:0]
			isPrefix = false
		}
	}

	keystring = strings.Join(keyBuffer, "")

	return
}

// DiscardTo discards and returns all pending input up to and including the provided targetKey
// If the targetKey is not present then all pending input is discarded
func (inputBuffer *InputBuffer) DiscardTo(targetKey string) (discarded string, targetKeyFound bool) {
	if !inputBuffer.hasInput() {
		return
	}

	var keyIndex int
	var key string

	for keyIndex, key = range inputBuffer.buffer {
		if key == targetKey {
			targetKeyFound = true
			break
		}
	}

	discarded = strings.Join(inputBuffer.buffer[:keyIndex+1], "")
	inputBuffer.buffer = inputBuffer.buffer[keyIndex+1:]

	return
}
