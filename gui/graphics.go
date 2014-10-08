package gui

import (
	"github.com/andlabs/ui"
	"image"
)

var area ui.Area
var window ui.Window

func mainGraphics(title string, handler Interface) error {
	go goGraphics(title, handler)

	return ui.Go()
}

func goGraphics(title string, handler Interface) {
	const w, h = 800, 600
	area = ui.NewArea(w, h, &graphicsHandler{})
	window = ui.NewWindow(title, w, h, area)

	window.OnClosing(func() bool {
		return handler.Closing()
	})

	<-exitch
	ui.Stop()
}

type graphicsHandler struct{ h Interface }

func (g *graphicsHandler) Paint(cliprect image.Rectangle) *image.RGBA {
	img := image.NewRGBA(cliprect)

	w, h := g.h.SpriteSize()

	for x := cliprect.Min.X / w; x <= cliprect.Max.X/w; x++ {
		for y := cliprect.Min.Y / h; y <= cliprect.Max.Y/h; y++ {
			s := g.h.SpriteAt(x, y)
			// TODO
			_ = s
		}
	}

	return img
}

func (g *graphicsHandler) Mouse(e ui.MouseEvent) {
	// TODO
}

func (g *graphicsHandler) Key(e ui.KeyEvent) (handled bool) {
	// TODO

	return false
}
