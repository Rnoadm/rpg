package main

import (
	"errors"
	"github.com/Rnoadm/rpg"
)

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

func (p *Pickaxe) Durability() int {
	return int(p.d)
}

func (p *Pickaxe) GobEncode() (data []byte, err error) {
	data = append(data, p.d)
	return
}
func (p *Pickaxe) GobDecode(data []byte) (err error) {
	p.d = data[0]
	return
}
