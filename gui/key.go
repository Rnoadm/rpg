package gui

import (
	"github.com/andlabs/ui"
	"github.com/nsf/termbox-go"
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
	PageUp
	PageDown
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
	termbox.KeyPgup:      PageUp,
	termbox.KeyPgdn:      PageDown,
}

var graphicsKeys = map[ui.ExtKey]Key{
	ui.F1:  F1,
	ui.F2:  F2,
	ui.F3:  F3,
	ui.F4:  F4,
	ui.F5:  F5,
	ui.F6:  F6,
	ui.F7:  F7,
	ui.F8:  F8,
	ui.F9:  F9,
	ui.F10: F10,
	ui.F11: F11,
	ui.F12: F12,

	ui.Up:    Up,
	ui.Down:  Down,
	ui.Left:  Left,
	ui.Right: Right,

	ui.Escape: Escape,
	// Backspace is a special case handled in code.
	ui.Insert:   Insert,
	ui.Delete:   Delete,
	ui.Home:     Home,
	ui.End:      End,
	ui.PageUp:   PageUp,
	ui.PageDown: PageDown,
}
