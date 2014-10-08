package gui

import (
	"github.com/andlabs/ui"
	"image"
	"image/draw"
)

var area ui.Area
var window ui.Window

func mainGraphics(title string, handler Interface) error {
	go goGraphics(title, handler)

	return ui.Go()
}

func goGraphics(title string, handler Interface) {
	ui.Do(func() {
		w, h := handler.SpriteSize()
		w *= 80
		h *= 25
		area = ui.NewArea(w, h, &graphicsHandler{handler, w, h})
		window = ui.NewWindow(title, w, h, area)

		window.OnClosing(func() bool {
			if handler.Closing() {
				select {
				case <-exitch:
				default:
					close(exitch)
				}
				return true
			}
			return false
		})

		window.Show()
	})

	for {
		select {
		case <-redrawch:
			ui.Do(area.RepaintAll)

		case <-exitch:
			ui.Stop()
			return
		}
	}
}

type graphicsHandler struct {
	handler Interface
	w, h    int
}

func (g *graphicsHandler) Paint(cliprect image.Rectangle) *image.RGBA {
	img := image.NewRGBA(cliprect)

	w, h := g.handler.SpriteSize()
	r := image.Rect(0, 0, w, h)

	g.w = (cliprect.Max.X + w - 1) / w
	g.h = (cliprect.Max.Y + h - 1) / h

	for x := cliprect.Min.X / w; x <= cliprect.Max.X/w; x++ {
		for y := cliprect.Min.Y / h; y <= cliprect.Max.Y/h; y++ {
			s := g.handler.SpriteAt(x, y, (cliprect.Max.X+w-1)/w, (cliprect.Max.Y+h-1)/h)
			for _, i := range s.Images {
				draw.Draw(img, r.Add(image.Pt(x*w, y*h)), i, i.Bounds().Min, draw.Over)
			}
		}
	}

	return img
}

func (g *graphicsHandler) Mouse(e ui.MouseEvent) {
	defer area.RepaintAll()

	x, y := e.Pos.X, e.Pos.Y
	w, h := g.handler.SpriteSize()
	x /= w
	y /= h

	g.handler.Mouse(x, y, (g.w+w-1)/w, (g.h+h-1)/h)
}

func (g *graphicsHandler) Key(e ui.KeyEvent) (handled bool) {
	defer area.RepaintAll()

	if e.Up {
		return false
	}
	if e.Modifier != 0 {
		return false
	}
	if e.Modifiers&^ui.Shift != 0 {
		return false
	}

	if e.Key == '\b' {
		return g.handler.Key(Backspace)
	}
	if e.Key == 0 {
		if k, ok := graphicsKeys[e.ExtKey]; ok {
			return g.handler.Key(k)
		}
	}

	if e.Key != 0 && e.Modifiers&ui.Shift == ui.Shift {
		if k, ok := graphicsShift[e.Key]; ok {
			return g.handler.Rune(k)
		}
	}
	if k := e.EffectiveKey(); k != 0 {
		return g.handler.Rune(rune(k))
	}
	return false
}

var graphicsShift = map[byte]rune{
	'`':  '~',
	'1':  '!',
	'2':  '@',
	'3':  '#',
	'4':  '$',
	'5':  '%',
	'6':  '^',
	'7':  '&',
	'8':  '*',
	'9':  '(',
	'0':  ')',
	'-':  '_',
	'=':  '+',
	'q':  'Q',
	'w':  'W',
	'e':  'E',
	'r':  'R',
	't':  'T',
	'y':  'Y',
	'u':  'U',
	'i':  'I',
	'o':  'O',
	'p':  'P',
	'[':  '{',
	']':  '}',
	'\\': '|',
	'a':  'A',
	's':  'S',
	'd':  'D',
	'f':  'F',
	'g':  'G',
	'h':  'H',
	'j':  'J',
	'k':  'K',
	'l':  'L',
	';':  ':',
	'\'': '"',
	'z':  'Z',
	'x':  'X',
	'c':  'C',
	'v':  'V',
	'b':  'B',
	'n':  'N',
	'm':  'M',
	',':  '<',
	'.':  '>',
	'/':  '?',
}
