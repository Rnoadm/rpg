package main

import (
	"github.com/Rnoadm/rpg"
	"github.com/Rnoadm/rpg/gui"
	"github.com/Rnoadm/rpg/history"
	"image"
	"io"
	"strconv"
)

type Handler struct {
	h              *history.History
	s              *rpg.State
	playerSprite   *image.RGBA
	fontSprites    *image.RGBA
	terrainSprites *image.RGBA
	pickaxeCount   *image.RGBA
	nextFrame      <-chan struct{}
	replayDone     chan<- struct{}
}

func (v *Handler) Closing() bool {
	return true
}

func (v *Handler) PreRender(w, h int) {
	if *flagReplay > 0 {
		s, err := v.h.Seek(1, history.SeekCur)
		if err == io.EOF {
			*flagReplay = 0
			close(v.replayDone)
		} else if err != nil {
			panic(err)
		} else {
			v.s = s
		}
	}

}

func (v *Handler) SpriteAt(x, y, w, h int) (sprite *gui.Sprite) {
	w2, h2 := w/2, h/2

	v.s.Atomic(func(s *rpg.State) bool {
		player := s.Get(s.ByComponent(PlayerType)[0])

		center := player.Component(rpg.LocationType).(*rpg.Location)

		m := player.Component(rpg.MessagesType).(*rpg.Messages)

		minedLocations := s.Get(s.ByComponent(MinedLocationsType)[0]).Component(MinedLocationsType).(*MinedLocations)

		ex, ey, ez := center.Get()

		ex += int64(x - w2)
		ey += int64(y - h2)

		sprite = &gui.Sprite{}

		if minedLocations.Has(ex, ey, ez) {
			sprite.Images = append(sprite.Images, v.terrainSprites.SubImage(image.Rect(int(((ex%3)+3)%3)*16, 0, int(((ex%3)+3)%3)*16+16, 16)))
			sprite.Rune = ' '
			sprite.Fg = gui.ColorBlack
			sprite.Bg = gui.ColorBlack
		} else {
			sprite.Images = append(sprite.Images, v.terrainSprites.SubImage(image.Rect(int((((ex+ey+ez)%3)+3)%3)*16, 16, int((((ex+ey+ez)%3)+3)%3)*16+16, 32)))
			sprite.Rune = '█'
			sprite.Fg = gui.ColorYellow
			sprite.Bg = gui.ColorYellow
		}
		if w2 == x && h2 == y {
			sprite.Images = append(sprite.Images, v.playerSprite)
			sprite.Rune = '⁈'
			sprite.Fg = gui.ColorRed
			sprite.Bg = gui.ColorBlack
		}
		if h-1 == y && m.Len() != 0 && m.At(m.Len()-1).Time == v.h.Tell() {
			msg := []rune(m.At(m.Len() - 1).Text)
			if x != 0 && x <= len(msg) {
				if msg[x-1] >= 'a' && msg[x-1] <= 'z' {
					sprite.Images = append(sprite.Images, v.fontSprites.SubImage(image.Rect(int(msg[x-1]-'a'+1)*16, 0, int(msg[x-1]-'a'+2)*16, 16)))
				} else {
					sprite.Images = append(sprite.Images, v.fontSprites)
				}
				sprite.Rune = msg[x-1]
			} else {
				sprite.Rune = ' '
				sprite.Images = append(sprite.Images, v.fontSprites)
			}
			sprite.Fg = gui.ColorBrightRed
			sprite.Bg = gui.ColorBlack
		}
		if h-1 == y {
			d := 0
			for _, o := range player.Component(rpg.ContainerType).(*rpg.Container).ByComponent(PickaxeType) {
				d += o.Component(PickaxeType).(*Pickaxe).Durability()
			}
			charges := strconv.Itoa(d)
			if x+len(charges)-w >= -1 {
				if x+len(charges)-w == -1 {
					sprite.Images = append(sprite.Images, v.pickaxeCount)
					sprite.Rune = '×'
					sprite.Fg = gui.ColorWhite
					sprite.Bg = gui.ColorBlack
				} else {
					c := charges[x+len(charges)-w]
					sprite.Images = append(sprite.Images, v.pickaxeCount.SubImage(image.Rect(int(c-'0'+1)*16, 0, int(c-'0'+2)*16, 16)))
					sprite.Rune = rune(c)
					sprite.Fg = gui.ColorBrightWhite
					sprite.Bg = gui.ColorBlack
				}
			}
		}
		return false
	})

	return
}

func (v *Handler) SpriteSize() (w, h int) {
	return 16, 16
}

func (v *Handler) Mouse(x, y, w, h int) {
	v.moveCharacter(int64(x-w/2), int64(y-h/2))
}

func (v *Handler) Rune(r rune) (handled bool) {
	switch r {
	case 'p':
		v.s.Atomic(func(s *rpg.State) bool {
			player := s.Get(s.ByComponent(PlayerType)[0])
			inventory := player.Component(rpg.ContainerType).(*rpg.Container)
			ores := inventory.ByComponent(OreType)
			if len(ores) == 0 {
				player.Component(rpg.MessagesType).(*rpg.Messages).Append(rpg.Message{
					Kind:   "error",
					Source: player.ID(),
					Text:   "no ores available",
					Time:   v.h.Tell() + 1,
				})
				return true
			}

			ore := ores[0]
			inventory.Remove(ore)
			s.Delete(ore.ID())
			_, pickaxe := s.Create(PickaxeFactory, rpg.LocationFactory)
			inventory.Add(pickaxe)

			return true
		})
		v.h.Append(v.s)
		return true
	}
	return false
}

func (v *Handler) Key(k gui.Key) (handled bool) {
	switch k {
	case gui.Up:
		v.moveCharacter(0, -1)
		return true
	case gui.Down:
		v.moveCharacter(0, 1)
		return true
	case gui.Left:
		v.moveCharacter(-1, 0)
		return true
	case gui.Right:
		v.moveCharacter(1, 0)
		return true
	}

	return false
}

func (v *Handler) moveCharacter(dx, dy int64) {
	if *flagReplay > 0 {
		return
	}

	v.s.Atomic(func(s *rpg.State) bool {
		player := s.Get(s.ByComponent(PlayerType)[0])

		center := player.Component(rpg.LocationType).(*rpg.Location)

		minedLocations := s.Get(s.ByComponent(MinedLocationsType)[0]).Component(MinedLocationsType).(*MinedLocations)

		x, y, z := center.Get()
		x += dx
		y += dy
		if !minedLocations.Has(x, y, z) {
			container := player.Component(rpg.ContainerType).(*rpg.Container)
			msg := &rpg.Message{
				Kind:   "error",
				Source: player.ID(),
				Time:   v.h.Tell() + 1,
				Text:   "no pickaxe in inventory",
			}
			for _, item := range container.ByComponent(PickaxeType) {
				p := item.Component(PickaxeType).(*Pickaxe)
				_, o, err := p.Use(x, y, z)
				if err == nil {
					container.Add(o)
					msg = nil
					break
				}
				msg = &rpg.Message{
					Kind:   "error",
					Source: item.ID(),
					Time:   v.h.Tell() + 1,
					Text:   err.Error(),
				}
			}
			if msg != nil {
				player.Component(rpg.MessagesType).(*rpg.Messages).Append(*msg)
			}
		} else {
			if center.Dist(x, y, z) == 1*1 {
				center.Set(x, y, z)
			}
		}
		return true
	})
	v.h.Append(v.s)
}
