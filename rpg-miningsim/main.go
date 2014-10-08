package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	"github.com/Rnoadm/rpg"
	"github.com/Rnoadm/rpg/gui"
	"github.com/Rnoadm/rpg/history"
	"image"
	"image/png"
	"io"
	"os"
	"sort"
	"time"
)

var (
	flagFilename = flag.String("f", "miningsim.sav", "filename for save file")
	flagReplay   = flag.Duration("replay", 0, "play back the game up to this point with this delay between frames")
)

func init() {
	rpg.MaxMessages = 1
}

type Player struct{}

func PlayerFactory(o *rpg.Object) rpg.Component {
	return &Player{}
}

var PlayerType = rpg.RegisterComponent(PlayerFactory)

func (p *Player) Clone(o *rpg.Object) rpg.Component {
	return &Player{}
}

func (p *Player) GobEncode() ([]byte, error) { return nil, nil }
func (p *Player) GobDecode([]byte) error     { return nil }

type Ore struct{}

func OreFactory(o *rpg.Object) rpg.Component {
	return &Ore{}
}

var OreType = rpg.RegisterComponent(OreFactory)

func (p *Ore) Clone(o *rpg.Object) rpg.Component {
	return &Ore{}
}

func (p *Ore) GobEncode() ([]byte, error) { return nil, nil }
func (p *Ore) GobDecode([]byte) error     { return nil }

type MinedLocations struct {
	l map[[3]int64]bool
	o *rpg.Object
}

func MinedLocationsFactory(o *rpg.Object) rpg.Component {
	return &MinedLocations{l: make(map[[3]int64]bool), o: o}
}

var MinedLocationsType = rpg.RegisterComponent(MinedLocationsFactory)

func (m *MinedLocations) Clone(o *rpg.Object) rpg.Component {
	l := make(map[[3]int64]bool, len(m.l))
	for loc := range m.l {
		l[loc] = true
	}
	return &MinedLocations{l: l, o: o}
}

func (m *MinedLocations) Has(x, y, z int64) bool {
	return m.l[[3]int64{x, y, z}]
}

func (m *MinedLocations) Add(x, y, z int64) {
	m.l[[3]int64{x, y, z}] = true
	m.o.Modified()
}

type sortLocations [][3]int64

func (l sortLocations) Len() int      { return len(l) }
func (l sortLocations) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l sortLocations) Less(i, j int) bool {
	for k := 0; k < 3; k++ {
		if l[i][k] < l[j][k] {
			return true
		}
		if l[i][k] > l[j][k] {
			return false
		}
	}
	return false
}

func (m *MinedLocations) GobEncode() (data []byte, err error) {
	var l sortLocations
	for loc := range m.l {
		l = append(l, loc)
	}
	sort.Sort(l)

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(&l)
	data = buf.Bytes()
	return
}
func (m *MinedLocations) GobDecode(data []byte) (err error) {
	var l sortLocations
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&l)
	if err != nil {
		return
	}

	for _, loc := range l {
		m.l[loc] = true
	}

	return
}

type Pickaxe struct {
	d uint8
	o *rpg.Object
}

func PickaxeFactory(o *rpg.Object) rpg.Component {
	return &Pickaxe{d: 10, o: o}
}

var PickaxeType = rpg.RegisterComponent(PickaxeFactory)

func (p *Pickaxe) Clone(o *rpg.Object) rpg.Component {
	return &Pickaxe{d: p.d, o: o}
}

var (
	ErrCantReach     = errors.New("cannot reach target")
	ErrPickaxeBroken = errors.New("pickaxe is broken")
	ErrNoOreThere    = errors.New("no ore at target location")
)

func (p *Pickaxe) Use(x, y, z int64) (rpg.ObjectIndex, *rpg.Object, error) {
	if p.d == 0 {
		return 0, nil, ErrPickaxeBroken
	}
	if p.o.Component(rpg.LocationType).(*rpg.Location).Dist(x, y, z) > 1*1 {
		return 0, nil, ErrCantReach
	}

	m := p.o.State().Get(p.o.State().ByComponent(MinedLocationsType)[0]).Component(MinedLocationsType).(*MinedLocations)
	if m.Has(x, y, z) {
		return 0, nil, ErrNoOreThere
	}

	id, o := p.o.State().Create(OreFactory)
	m.Add(x, y, z)
	p.d--
	p.o.Modified()
	return id, o, nil
}

func (p *Pickaxe) GobEncode() (data []byte, err error) {
	data = append(data, p.d)
	return
}
func (p *Pickaxe) GobDecode(data []byte) (err error) {
	p.d = data[0]
	return
}

func main() {
	flag.Parse()

	f, err := os.OpenFile(*flagFilename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	h := history.NewHistory(f)
	s, err := h.Seek(0, history.SeekEnd)
	if err == io.EOF {
		s = rpg.NewState()
		s.Atomic(func(s *rpg.State) bool {
			_, o := s.Create(MinedLocationsFactory)
			for x := int64(-2); x <= 2; x++ {
				for y := int64(-2); y <= 2; y++ {
					for z := int64(0); z <= 0; z++ {
						o.Component(MinedLocationsType).(*MinedLocations).Add(x, y, z)
					}
				}
			}
			_, o = s.Create(PlayerFactory, rpg.ContainerFactory, rpg.LocationFactory, rpg.MessagesFactory)
			_, pickaxe := s.Create(PickaxeFactory, rpg.LocationFactory)
			o.Component(rpg.ContainerType).(*rpg.Container).Add(pickaxe)
			return true
		})
		err = h.Append(s)
	}
	if err != nil {
		panic(err)
	}

	nextFrame := make(chan struct{}, 1)
	replayDone := make(chan struct{})
	if *flagReplay > 0 {
		h.Reset()
		go func(d time.Duration) {
			for {
				select {
				case nextFrame <- struct{}{}:
				default:
				}
				gui.Redraw()
				select {
				case <-replayDone:
					return
				default:
				}
				time.Sleep(d)
			}
		}(*flagReplay)
	}

	playerSprite, err := png.Decode(bytes.NewReader(MiningsimPlayerPng))
	if err != nil {
		panic(err)
	}

	fontSprites, err := png.Decode(bytes.NewReader(MiningsimFontPng))
	if err != nil {
		panic(err)
	}

	terrainSprites, err := png.Decode(bytes.NewReader(MiningsimTerrainPng))
	if err != nil {
		panic(err)
	}

	gui.Main("Mining Simulator 2014", &Handler{
		h:              h,
		s:              s,
		playerSprite:   playerSprite,
		fontSprites:    fontSprites.(*image.Paletted),
		terrainSprites: terrainSprites.(*image.Paletted),
		nextFrame:      nextFrame,
		replayDone:     replayDone,
	})
}

type Handler struct {
	h              *history.History
	s              *rpg.State
	playerSprite   image.Image
	fontSprites    *image.Paletted
	terrainSprites *image.Paletted
	nextFrame      <-chan struct{}
	replayDone     chan<- struct{}
}

func (v *Handler) Closing() bool {
	return true
}

func (v *Handler) SpriteAt(x, y, w, h int) (sprite *gui.Sprite) {
	if *flagReplay > 0 {
		select {
		case <-v.nextFrame:
			s, err := v.h.Seek(1, history.SeekCur)
			if err == io.EOF {
				*flagReplay = 0
				close(v.replayDone)
			} else if err != nil {
				panic(err)
			} else {
				v.s = s
			}
		default:
		}
	}

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
			var msg *rpg.Message
			for _, item := range container.ByComponent(PickaxeType) {
				p := item.Component(PickaxeType).(*Pickaxe)
				_, o, err := p.Use(x, y, z)
				if err == nil {
					container.Add(o)
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
