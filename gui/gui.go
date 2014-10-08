// Package gui provides a text/graphical rendering system for tile-based games.
package gui

import "flag"

var flagText = flag.Bool("text", false, "Use text mode instead of graphics mode.")

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

func Exit() {
	close(exitch)
}

var redrawch = make(chan struct{}, 1)

func Redraw() {
	select {
	case redrawch <- struct{}{}:
	default:
	}
}

type Interface interface {
	SpriteSize() (w, h int)
	SpriteAt(x, y, w, h int) *Sprite
	Mouse(x, y, w, h int)
	Rune(r rune) (handled bool)
	Key(k Key) (handled bool)
	Closing() (allow bool)
}
