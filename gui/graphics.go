package gui

import (
	"github.com/skelterjohn/go.wde"
	_ "github.com/skelterjohn/go.wde/init"
	"image"
	"image/draw"
)

func mainGraphics(title string, handler Interface) error {
	errch := make(chan error, 2)
	go goGraphics(title, handler, errch)

	wde.Run()

	return <-errch
}

func goGraphics(title string, handler Interface, errch chan<- error) {
	defer func() {
		errch <- nil
	}()
	defer wde.Stop()

	w, h := handler.SpriteSize()
	w *= 80
	h *= 25
	window, err := wde.NewWindow(w, h)
	if err != nil {
		errch <- err
		return
	}
	window.SetTitle(title)
	window.Show()

	eventch := window.EventChan()

	for {
		select {
		case event := <-eventch:
			switch e := event.(type) {
			case wde.CloseEvent:
				if handler.Closing() {
					return
				} else {
					window.Show()
				}

			case wde.ResizeEvent:
				Redraw()

			case wde.MouseDraggedEvent, wde.MouseEnteredEvent, wde.MouseExitedEvent, wde.MouseMovedEvent, wde.MouseUpEvent:
				// ignore

			case wde.MouseDownEvent:
				if e.Which&wde.LeftButton == 0 {
					continue
				}
				x, y := e.Where.X, e.Where.Y
				w, h := handler.SpriteSize()
				ww, wh := window.Size()
				handler.Mouse(x/w, y/h, ww/w, wh/h)
				Redraw()

			case wde.KeyDownEvent, wde.KeyUpEvent:
				// ignore

			case wde.KeyTypedEvent:
				if k, ok := graphicsKeys[e.Key]; ok {
					_ = handler.Key(k)
				} else {
					for _, r := range e.Glyph {
						_ = handler.Rune(r)
					}
				}
				Redraw()
			}

		case <-redrawch:
			s := window.Screen()
			r := s.Bounds()
			s.CopyRGBA(graphicsPaint(handler, r), r)
			window.FlushImage(r)

		case <-exitch:
			return
		}
	}
}

func graphicsPaint(handler Interface, cliprect image.Rectangle) *image.RGBA {
	img := image.NewRGBA(cliprect)

	w, h := handler.SpriteSize()
	r := image.Rect(0, 0, w, h)
	ww, wh := cliprect.Max.X/w, cliprect.Max.Y/h

	handler.PreRender(ww, wh)

	for x := cliprect.Min.X / w; x < ww; x++ {
		for y := cliprect.Min.Y / h; y < wh; y++ {
			s := handler.SpriteAt(x, y, ww, wh)
			for _, i := range s.Images {
				draw.Draw(img, r.Add(image.Pt(x*w, y*h)), i, i.Bounds().Min, draw.Over)
			}
		}
	}

	return img
}
