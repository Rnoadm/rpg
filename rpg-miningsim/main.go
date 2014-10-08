package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/Rnoadm/rpg"
	"github.com/Rnoadm/rpg/gui"
	"github.com/Rnoadm/rpg/history"
	"image"
	"image/png"
	"io"
	"os"
	"sort"
)

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
	Durability uint8
	o          *rpg.Object
}

func PickaxeFactory(o *rpg.Object) rpg.Component {
	return &Pickaxe{Durability: 10, o: o}
}

var PickaxeType = rpg.RegisterComponent(PickaxeFactory)

func (p *Pickaxe) Clone(o *rpg.Object) rpg.Component {
	return &Pickaxe{Durability: p.Durability, o: o}
}

var (
	ErrCantReach     = errors.New("cannot reach target")
	ErrPickaxeBroken = errors.New("pickaxe is broken")
	ErrNoOreThere    = errors.New("no ore at target location")
)

func (p *Pickaxe) Use(l *rpg.Location) (rpg.ObjectIndex, *rpg.Object, error) {
	if p.Durability == 0 {
		return 0, nil, ErrPickaxeBroken
	}
	if p.o.Component(rpg.LocationType).(*rpg.Location).Dist(l) > 1*1 {
		return 0, nil, ErrCantReach
	}

	m := p.o.State().Get(p.o.State().ByComponent(MinedLocationsType)[0]).Component(MinedLocationsType).(*MinedLocations)
	if m.Has(l.Get()) {
		return 0, nil, ErrNoOreThere
	}

	id, o := p.o.State().Create(OreFactory)
	m.Add(l.Get())
	p.Durability--
	p.o.Modified()
	return id, o, nil
}

func main() {
	f, err := os.OpenFile("miningsim.sav", os.O_CREATE|os.O_RDWR, 0644)
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

	playerSprite, err := png.Decode(bytes.NewReader(MiningsimPlayerPng))
	if err != nil {
		panic(err)
	}

	gui.Main("Mining Simulator 2014", &Handler{h: h, s: s, playerSprite: playerSprite})
}

type Handler struct {
	h            *history.History
	s            *rpg.State
	playerSprite image.Image
}

func (v *Handler) Closing() bool {
	return true
}

func (v *Handler) SpriteAt(x, y, w, h int) (sprite *gui.Sprite) {
	w2, h2 := w/2, h/2

	v.s.Atomic(func(s *rpg.State) bool {
		player := s.Get(s.ByComponent(PlayerType)[0])

		center := player.Component(rpg.LocationType).(*rpg.Location)

		minedLocations := s.Get(s.ByComponent(MinedLocationsType)[0]).Component(MinedLocationsType).(*MinedLocations)

		ex, ey, ez := center.Get()

		ex += int64(x - w2)
		ey += int64(y - h2)

		sprite = &gui.Sprite{}

		if w2 == x && h2 == y {
			sprite.Images = append(sprite.Images, v.playerSprite)
			sprite.Rune = '⁈'
			sprite.Fg = gui.ColorRed
			sprite.Bg = gui.ColorBlack
		} else if h-1 == y {
			// TODO: text
			sprite.Rune = ' '
			sprite.Fg = gui.ColorBlack
			sprite.Bg = gui.ColorBlack
		} else if minedLocations.Has(ex, ey, ez) {
			sprite.Rune = ' '
			sprite.Fg = gui.ColorBlack
			sprite.Bg = gui.ColorBlack
		} else {
			sprite.Rune = '█'
			sprite.Fg = gui.ColorYellow
			sprite.Bg = gui.ColorYellow
		}

		return false
	})

	return
}

func (v *Handler) SpriteSize() (w, h int) {
	return 16, 16
}
