package main

import "github.com/Rnoadm/rpg"

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
