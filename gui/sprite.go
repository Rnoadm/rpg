package gui

import (
	"github.com/nsf/termbox-go"
	"image"
)

type Color termbox.Attribute

const (
	ColorBlack   = Color(termbox.ColorBlack)
	ColorBlue    = Color(termbox.ColorBlue)
	ColorCyan    = Color(termbox.ColorCyan)
	ColorGreen   = Color(termbox.ColorGreen)
	ColorMagenta = Color(termbox.ColorMagenta)
	ColorRed     = Color(termbox.ColorRed)
	ColorWhite   = Color(termbox.ColorWhite)
	ColorYellow  = Color(termbox.ColorYellow)

	ColorBrightBlack   = Color(termbox.ColorBlack | termbox.AttrBold)
	ColorBrightBlue    = Color(termbox.ColorBlue | termbox.AttrBold)
	ColorBrightCyan    = Color(termbox.ColorCyan | termbox.AttrBold)
	ColorBrightGreen   = Color(termbox.ColorGreen | termbox.AttrBold)
	ColorBrightMagenta = Color(termbox.ColorMagenta | termbox.AttrBold)
	ColorBrightRed     = Color(termbox.ColorRed | termbox.AttrBold)
	ColorBrightWhite   = Color(termbox.ColorWhite | termbox.AttrBold)
	ColorBrightYellow  = Color(termbox.ColorYellow | termbox.AttrBold)
)

type Sprite struct {
	Images []image.Image
	Rune   rune
	Fg, Bg Color
}
