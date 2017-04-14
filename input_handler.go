package main

type InputHandler struct {
	buffer      []rune
	keyBindings KeyBindings
}

func NewInputHandler(keyBindings KeyBindings) *InputHandler {
	return &InputHandler{
		buffer:      make([]rune, 0),
		keyBindings: keyBindings,
	}
}

func (inputHandler *InputHandler) Append(input string) {
	inputHandler.buffer = append(inputHandler.buffer, []rune(input)...)
}

func (inputHandler *InputHandler) prepend(keystring string) {
	inputHandler.buffer = append([]rune(keystring), inputHandler.buffer...)
}

func (inputHandler *InputHandler) pop() (char rune) {
	char = inputHandler.buffer[0]
	inputHandler.buffer = inputHandler.buffer[1:]
	return
}

func (inputHandler *InputHandler) hasInput() bool {
	return len(inputHandler.buffer) > 0
}

func (inputHandler *InputHandler) Process(viewHierarchy ViewHierarchy) (action Action, keystring string) {
	if !inputHandler.hasInput() {
		return
	}

	keyBuffer := make([]rune, 0)
	keyBindings := inputHandler.keyBindings
	isPrefix := false

OuterLoop:
	for inputHandler.hasInput() {
		keyBuffer = append(keyBuffer, inputHandler.pop())
		binding, prefix := keyBindings.Binding(viewHierarchy, string(keyBuffer))

		switch {
		case prefix:
			if len(inputHandler.buffer) == 0 {
				inputHandler.prepend(string(keyBuffer))
				return
			} else {
				isPrefix = true
			}
		case binding.bindingType == BT_ACTION:
			if binding.action != ACTION_NONE {
				action = binding.action
			} else if isPrefix {
				inputHandler.prepend(string(keyBuffer[1:]))
				keyBuffer = keyBuffer[0:1]
			}

			break OuterLoop
		case binding.bindingType == BT_KEYSTRING:
			inputHandler.prepend(binding.keystring)
			keyBuffer = keyBuffer[0:0]
			isPrefix = false
		}
	}

	keystring = string(keyBuffer)

	return
}
