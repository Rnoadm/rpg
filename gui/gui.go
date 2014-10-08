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

type Interface interface {
	SpriteSize() (w, h int)
	SpriteAt(x, y int) *Sprite
	Closing() (allow bool)
}
