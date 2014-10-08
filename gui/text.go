package gui

import "github.com/nsf/termbox-go"

func mainText(handler Interface) error {
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	termbox.HideCursor()

	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

	for {
		paintText(handler)

		select {
		case event := <-events:
			switch event.Type {
			case termbox.EventError:
				return event.Err

			case termbox.EventResize:
				// we repaint on the next iteration

			case termbox.EventMouse:
				w, h := termbox.Size()
				handler.Mouse(event.MouseX, event.MouseY, w, h)

			case termbox.EventKey:
				if event.Ch == 0 && event.Key == termbox.KeyEnter {
					event.Ch = '\n'
				}
				if event.Ch == 0 {
					if k, ok := textKeys[event.Key]; ok {
						if handler.Key(k) {
							continue
						}
					}

					if event.Key == termbox.KeyEsc {
						if handler.Closing() {
							return nil
						}
					}
				} else {
					handler.Rune(event.Ch)
				}
			}
		case <-redrawch:
			// we repaint on the next iteration

		case <-exitch:
			return nil
		}
	}
}

func paintText(handler Interface) {
	w, h := termbox.Size()
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			s := handler.SpriteAt(x, y, w, h)
			termbox.SetCell(x, y, s.Rune, termbox.Attribute(s.Fg), termbox.Attribute(s.Bg))
		}
	}
	termbox.Flush()
}
