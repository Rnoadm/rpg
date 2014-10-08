package gui

import "github.com/nsf/termbox-go"

func mainText(handler Interface) error {
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	termbox.HideCursor()

	for {
		paintText(handler)

		event := termbox.PollEvent()
		switch event.Type {
		case termbox.EventError:
			return event.Err

		case termbox.EventResize:
			// we repaint on the next iteration

		case termbox.EventMouse:
			// TODO

		case termbox.EventKey:
			if event.Ch == 0 {
				if event.Key == termbox.KeyEsc {
					if handler.Closing() {
						return nil
					}
				}

				// TODO
			} else {
				// TODO
			}
		}

		select {
		case <-exitch:
			return nil

		default:
		}
	}
}

func paintText(h Interface) {

}
