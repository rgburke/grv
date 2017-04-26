package main

import (
	"strings"
)

type InputBuffer struct {
	buffer      []string
	keyBindings KeyBindings
}

func NewInputBuffer(keyBindings KeyBindings) *InputBuffer {
	return &InputBuffer{
		buffer:      make([]string, 0),
		keyBindings: keyBindings,
	}
}

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
			} else {
				isPrefix = true
			}
		case binding.bindingType == BT_ACTION:
			if binding.action != ACTION_NONE {
				action = binding.action
			} else if isPrefix {
				inputBuffer.prepend(keyBuffer[1:])
				keyBuffer = keyBuffer[0:1]
			}

			break OuterLoop
		case binding.bindingType == BT_KEYSTRING:
			inputBuffer.prepend(TokeniseKeys(binding.keystring))
			keyBuffer = keyBuffer[0:0]
			isPrefix = false
		}
	}

	keystring = strings.Join(keyBuffer, "")

	return
}
