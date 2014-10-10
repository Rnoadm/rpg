package gui

import (
	"github.com/nsf/termbox-go"
	"github.com/skelterjohn/go.wde"
)

type Key uint

const (
	_ Key = iota
	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12

	Up
	Down
	Left
	Right

	Escape
	Backspace
	Insert
	Delete
	Home
	End
)

var textKeys = map[termbox.Key]Key{
	termbox.KeyF1:  F1,
	termbox.KeyF2:  F2,
	termbox.KeyF3:  F3,
	termbox.KeyF4:  F4,
	termbox.KeyF5:  F5,
	termbox.KeyF6:  F6,
	termbox.KeyF7:  F7,
	termbox.KeyF8:  F8,
	termbox.KeyF9:  F9,
	termbox.KeyF10: F10,
	termbox.KeyF11: F11,
	termbox.KeyF12: F12,

	termbox.KeyArrowUp:    Up,
	termbox.KeyArrowDown:  Down,
	termbox.KeyArrowLeft:  Left,
	termbox.KeyArrowRight: Right,

	termbox.KeyEsc:       Escape,
	termbox.KeyBackspace: Backspace,
	termbox.KeyInsert:    Insert,
	termbox.KeyDelete:    Delete,
	termbox.KeyHome:      Home,
	termbox.KeyEnd:       End,
}

var graphicsKeys = map[string]Key{
	wde.KeyF1:  F1,
	wde.KeyF2:  F2,
	wde.KeyF3:  F3,
	wde.KeyF4:  F4,
	wde.KeyF5:  F5,
	wde.KeyF6:  F6,
	wde.KeyF7:  F7,
	wde.KeyF8:  F8,
	wde.KeyF9:  F9,
	wde.KeyF10: F10,
	wde.KeyF11: F11,
	wde.KeyF12: F12,

	wde.KeyUpArrow:    Up,
	wde.KeyDownArrow:  Down,
	wde.KeyLeftArrow:  Left,
	wde.KeyRightArrow: Right,

	wde.KeyEscape:    Escape,
	wde.KeyBackspace: Backspace,
	wde.KeyInsert:    Insert,
	wde.KeyDelete:    Delete,
	wde.KeyHome:      Home,
	wde.KeyEnd:       End,
}
