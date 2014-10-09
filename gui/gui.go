// Package gui provides a text/graphical rendering system for tile-based games.
package gui

import "flag"

var flagText = flag.Bool("text", false, "Use text mode instead of graphics mode.")

// Main runs the gui. It must be called from the main function, not from a goroutine. Main
// will not return until the gui exits with an error or Exit is called.
func Main(title string, handler Interface) (err error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *flagText {
		return mainText(handler)
	} else {
		return mainGraphics(title, handler)
	}
}

var exitch = make(chan struct{})

// Exit tells the program to exit, then returns immediately. Main will return at some point
// in the future. Calling Exit twice panics.
func Exit() {
	close(exitch)
}

var redrawch = make(chan struct{}, 1)

// Redraw returns immediately. Some time later, the gui will redraw itself. It is not
// required to call Redraw after an input event (Mouse, Rune, or Key).
func Redraw() {
	select {
	case redrawch <- struct{}{}:
	default:
	}
}

// Interface is the set of methods that are called by gui.
//
// Methods with x, y, w, h parameters fulful x ∈ [0, w) ∧ y ∈ [0, h).
type Interface interface {
	// SpriteSize returns the width and height of each tile.
	SpriteSize() (w, h int)

	// PreRender is called with the width and height in tiles of the screen before
	// SpriteAt is called during a frame.
	PreRender(w, h int)

	// SpriteAt returns the Sprite at (x, y) on the screen.
	SpriteAt(x, y, w, h int) *Sprite

	// Mouse is called when the user left clicks on the screen.
	Mouse(x, y, w, h int)

	// Rune is called when a character is entered on the keyboard. Return false if
	// the OS should handle the key event instead.
	Rune(r rune) (handled bool)

	// Key is called when a special key is pressed on the keyboard. Return false if
	// the OS should handle the key event instead. The default action for Escape is
	// to exit the program.
	Key(k Key) (handled bool)

	// Closing is called before the user exits the program. Return false to cancel.
	Closing() (allow bool)
}
