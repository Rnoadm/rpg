package main

import "github.com/Rnoadm/rpg"

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
