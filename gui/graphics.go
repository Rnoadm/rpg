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
	area = ui.NewArea(w, h, &graphicsHandler{handler, w, h})
	window = ui.NewWindow(title, w, h, area)

	window.OnClosing(func() bool {
		return handler.Closing()
	})

	<-exitch
	ui.Stop()
}

type graphicsHandler struct {
	handler Interface
	w, h    int
}

func (g *graphicsHandler) Paint(cliprect image.Rectangle) *image.RGBA {
	img := image.NewRGBA(cliprect)

	w, h := g.handler.SpriteSize()

	for x := cliprect.Min.X / w; x <= cliprect.Max.X/w; x++ {
		for y := cliprect.Min.Y / h; y <= cliprect.Max.Y/h; y++ {
			s := g.handler.SpriteAt(x, y, (g.w+w-1)/w, (g.h+h-1)/h)
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
