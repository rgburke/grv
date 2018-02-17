package main

import (
	"bytes"
	"fmt"
	"regexp"
	"unicode"

	log "github.com/Sirupsen/logrus"
	gc "github.com/rgburke/goncurses"
)

const (
	ikmEscapeKey = 0x1B
	ikmCtrlMask  = 0x1F
)

var keyMap = map[gc.Key]string{
	gc.KEY_TAB:       "<Tab>",
	gc.KEY_RETURN:    "<Enter>",
	gc.KEY_DOWN:      "<Down>",
	gc.KEY_UP:        "<Up>",
	gc.KEY_LEFT:      "<Left>",
	gc.KEY_RIGHT:     "<Right>",
	gc.KEY_HOME:      "<Home>",
	gc.KEY_BACKSPACE: "<Backspace>",
	gc.KEY_F1:        "<F1>",
	gc.KEY_F2:        "<F2>",
	gc.KEY_F3:        "<F3>",
	gc.KEY_F4:        "<F4>",
	gc.KEY_F5:        "<F5>",
	gc.KEY_F6:        "<F6>",
	gc.KEY_F7:        "<F7>",
	gc.KEY_F8:        "<F8>",
	gc.KEY_F9:        "<F9>",
	gc.KEY_F10:       "<F10>",
	gc.KEY_F11:       "<F11>",
	gc.KEY_F12:       "<F12>",
	gc.KEY_DL:        "<Dl>",
	gc.KEY_IL:        "<Il>",
	gc.KEY_DC:        "<Dc>",
	gc.KEY_IC:        "<Ic>",
	gc.KEY_EIC:       "<Eic>",
	gc.KEY_CLEAR:     "<Clear>",
	gc.KEY_EOS:       "<Eos>",
	gc.KEY_EOL:       "<Eol>",
	gc.KEY_SF:        "<Sf>",
	gc.KEY_SR:        "<Sr>",
	gc.KEY_PAGEDOWN:  "<PageDown>",
	gc.KEY_PAGEUP:    "<PageUp>",
	gc.KEY_STAB:      "<Stab>",
	gc.KEY_CTAB:      "<Ctab>",
	gc.KEY_CATAB:     "<Catab>",
	gc.KEY_ENTER:     "<Enter>",
	gc.KEY_PRINT:     "<Print>",
	gc.KEY_LL:        "<Ll>",
	gc.KEY_A1:        "<A1>",
	gc.KEY_A3:        "<A3>",
	gc.KEY_B2:        "<B2>",
	gc.KEY_C1:        "<C1>",
	gc.KEY_C3:        "<C3>",
	gc.KEY_BTAB:      "<S-Tab>",
	gc.KEY_BEG:       "<Beg>",
	gc.KEY_CANCEL:    "<Cancel>",
	gc.KEY_CLOSE:     "<Close>",
	gc.KEY_COMMAND:   "<Command>",
	gc.KEY_COPY:      "<Copy>",
	gc.KEY_CREATE:    "<Create>",
	gc.KEY_END:       "<End>",
	gc.KEY_EXIT:      "<Exit>",
	gc.KEY_FIND:      "<Find>",
	gc.KEY_HELP:      "<Help>",
	gc.KEY_MARK:      "<Mark>",
	gc.KEY_MESSAGE:   "<Message>",
	gc.KEY_MOVE:      "<Move>",
	gc.KEY_NEXT:      "<Next>",
	gc.KEY_OPEN:      "<Open>",
	gc.KEY_OPTIONS:   "<Options>",
	gc.KEY_PREVIOUS:  "<Previous>",
	gc.KEY_REDO:      "<Redo>",
	gc.KEY_REFERENCE: "<Reference>",
	gc.KEY_REFRESH:   "<Refresh>",
	gc.KEY_REPLACE:   "<Replace>",
	gc.KEY_RESTART:   "<Restart>",
	gc.KEY_RESUME:    "<Resume>",
	gc.KEY_SAVE:      "<Save>",
	gc.KEY_SBEG:      "<S-Beg>",
	gc.KEY_SCANCEL:   "<S-Cancel>",
	gc.KEY_SCOMMAND:  "<S-Command>",
	gc.KEY_SCOPY:     "<S-Copy>",
	gc.KEY_SCREATE:   "<S-Create>",
	gc.KEY_SDC:       "<S-Dc>",
	gc.KEY_SDL:       "<S-Dl>",
	gc.KEY_SELECT:    "<Select>",
	gc.KEY_SEND:      "<S-End>",
	gc.KEY_SEOL:      "<S-Eol>",
	gc.KEY_SEXIT:     "<S-Exit>",
	gc.KEY_SFIND:     "<S-Find>",
	gc.KEY_SHELP:     "<S-Help>",
	gc.KEY_SHOME:     "<S-Home>",
	gc.KEY_SIC:       "<S-Ic>",
	gc.KEY_SLEFT:     "<S-Left>",
	gc.KEY_SMESSAGE:  "<S-Message>",
	gc.KEY_SMOVE:     "<S-Move>",
	gc.KEY_SNEXT:     "<S-Next>",
	gc.KEY_SOPTIONS:  "<S-Options>",
	gc.KEY_SPREVIOUS: "<S-Previous>",
	gc.KEY_SPRINT:    "<S-Print>",
	gc.KEY_SREDO:     "<S-Redo>",
	gc.KEY_SREPLACE:  "<S-Replace>",
	gc.KEY_SRIGHT:    "<S-Right>",
	gc.KEY_SRSUME:    "<S-Rsume>",
	gc.KEY_SSAVE:     "<S-Save>",
	gc.KEY_SSUSPEND:  "<S-Suspend>",
	gc.KEY_SUNDO:     "<S-Undo>",
	gc.KEY_SUSPEND:   "<Suspend>",
	gc.KEY_UNDO:      "<Undo>",
	gc.KEY_MOUSE:     "<Mouse>",
	gc.KEY_RESIZE:    "<Resize>",
	gc.KEY_MAX:       "<Max>",
}

var ncursesSpecialKeys map[string]bool

func init() {
	ncursesSpecialKeys = make(map[string]bool, len(keyMap))

	for _, keyStr := range keyMap {
		ncursesSpecialKeys[keyStr] = true
	}
}

// InputKeyMapper maps ncurses characters to key string representations and groups byte sequences into UTF-8 characters
type InputKeyMapper struct {
	ui               InputUI
	char             bytes.Buffer
	expectedCharSize int
}

// NewInputKeyMapper creates a new instance
func NewInputKeyMapper(ui InputUI) *InputKeyMapper {
	return &InputKeyMapper{
		ui: ui,
	}
}

// GetKeyInput fetches the next character or key string returned from the UI
func (inputKeyMapper *InputKeyMapper) GetKeyInput() (key string, err error) {
	for {
		keyPressEvent, err := inputKeyMapper.ui.GetInput(false)
		mappedKey, isMappedKey := keyMap[gc.Key(keyPressEvent)]

		switch {
		case err != nil:
			return key, err
		case keyPressEvent == UINoKey:
			return key, err
		case inputKeyMapper.isProcessingUTF8Char():
			err = inputKeyMapper.processUTF8ContinuationByte(keyPressEvent)
		case isMappedKey:
			return mappedKey, err
		case keyPressEvent == ikmEscapeKey:
			return inputKeyMapper.metaKeyString(), err
		case isControlKey(keyPressEvent):
			return controlKeyString(keyPressEvent), err
		default:
			err = inputKeyMapper.processUTF8Char(keyPressEvent)
		}

		if err != nil {
			log.Errorf("Discarding input character: %v", err)
			inputKeyMapper.clearChar()
		} else if inputKeyMapper.isUTF8CharComplete() {
			return inputKeyMapper.getAndClearChar(), err
		}
	}
}

func (inputKeyMapper *InputKeyMapper) isProcessingUTF8Char() bool {
	return inputKeyMapper.expectedCharSize > 0
}

func (inputKeyMapper *InputKeyMapper) isUTF8CharComplete() bool {
	return inputKeyMapper.expectedCharSize > 0 && inputKeyMapper.char.Len() == inputKeyMapper.expectedCharSize
}

func (inputKeyMapper *InputKeyMapper) clearChar() {
	inputKeyMapper.expectedCharSize = 0
	inputKeyMapper.char.Reset()
}

func (inputKeyMapper *InputKeyMapper) getAndClearChar() (char string) {
	char = inputKeyMapper.char.String()
	inputKeyMapper.clearChar()
	return
}

func (inputKeyMapper *InputKeyMapper) processUTF8ContinuationByte(keyPressEvent Key) (err error) {
	if keyPressEvent>>6 != 0x02 {
		err = fmt.Errorf("Invalid UTF-8 continuation byte: %b", keyPressEvent)
	} else {
		inputKeyMapper.char.WriteByte(byte(keyPressEvent))
	}

	return
}

func (inputKeyMapper *InputKeyMapper) processUTF8Char(keyPressEvent Key) (err error) {
	switch {
	case keyPressEvent < 0x80:
		inputKeyMapper.expectedCharSize = 1
		inputKeyMapper.char.WriteByte(byte(keyPressEvent))
	case keyPressEvent>>5 == 0x06:
		inputKeyMapper.expectedCharSize = 2
		inputKeyMapper.char.WriteByte(byte(keyPressEvent))
	case keyPressEvent>>4 == 0x0E:
		inputKeyMapper.expectedCharSize = 3
		inputKeyMapper.char.WriteByte(byte(keyPressEvent))
	case keyPressEvent>>3 == 0x1E:
		inputKeyMapper.expectedCharSize = 4
		inputKeyMapper.char.WriteByte(byte(keyPressEvent))
	default:
		err = fmt.Errorf("Invalid UTF-8 starting byte: %b", keyPressEvent)
	}

	return
}

func (inputKeyMapper *InputKeyMapper) metaKeyString() string {
	keyPressEvent, err := inputKeyMapper.ui.GetInput(true)

	if err != nil || keyPressEvent == 0 {
		return "<Escape>"
	}

	return fmt.Sprintf("<M-%c>", keyPressEvent)
}

func isControlKey(keyPressEvent Key) bool {
	return keyPressEvent >= (ikmCtrlMask&'@') && keyPressEvent <= (ikmCtrlMask&'_')
}

func controlKeyString(keyPressEvent Key) string {
	return fmt.Sprintf("<C-%c>", unicode.ToLower(rune(keyPressEvent|0x40)))
}

// TokeniseKeys breaks a key string sequence down in to the individual keys is consists of
func TokeniseKeys(keysString string) (keys []string) {
	runes := []rune(keysString)

	for i := 0; i < len(runes); i++ {
		if runes[i] == '<' {
			if key, isSpecialKey := specialKeyString(string(runes[i:])); isSpecialKey {
				keys = append(keys, key)
				i += len(key) - 1
				continue
			}
		}

		keys = append(keys, string(runes[i]))
	}

	return
}

func specialKeyString(potentialKeyString string) (key string, isSpecialKey bool) {
	endIndex := -1

	for charIndex, char := range potentialKeyString {
		if char == '>' {
			endIndex = charIndex
			break
		}
	}

	if endIndex == -1 {
		return
	}

	keyString := potentialKeyString[0 : endIndex+1]

	matched, err := regexp.MatchString("^<(C|M)-.>$", keyString)
	if err != nil {
		log.Errorf("Invalid Regex: %v", err)
		return
	}

	if matched || isValidAction(keyString) || isNCursesSpecialKey(keyString) {
		key = keyString
		isSpecialKey = true
	}

	return
}

func isNCursesSpecialKey(key string) bool {
	_, ok := ncursesSpecialKeys[key]
	return ok
}
