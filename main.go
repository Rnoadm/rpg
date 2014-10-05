// +build ignore

package main

import (
	"fmt"
	"github.com/Rnoadm/rpg"
	"reflect"
)

type Name string

func NameFactory(name string) rpg.ComponentFactory {
	return Name(name).Clone
}

var NameType = reflect.TypeOf(Name(""))

func (n Name) Clone(*rpg.Object) rpg.Component {
	return n
}

func main() {
	global := rpg.NewState()

	var personA, personB, itemA, itemB rpg.ObjectIndex
	if !global.Atomic(func(s *rpg.State) bool {
		var pa, pb, ia, ib *rpg.Object
		personA, pa = s.Create(NameFactory("person A"), rpg.ContainerFactory)
		personB, pb = s.Create(NameFactory("person B"), rpg.ContainerFactory)
		itemA, ia = s.Create(NameFactory("item A"))
		itemB, ib = s.Create(NameFactory("item B"))

		if !pa.Component(rpg.ContainerType).(*rpg.Container).Add(ia) {
			panic("unreachable")
		}
		if !pb.Component(rpg.ContainerType).(*rpg.Container).Add(ib) {
			panic("unreachable")
		}
		return true
	}) {
		panic("unreachable")
	}

	pa := global.Get(personA)
	for _, o := range pa.Component(rpg.ContainerType).(*rpg.Container).Contents() {
		fmt.Println(pa.Component(NameType), "has", o.Component(NameType))
	}

	pb := global.Get(personB)
	for _, o := range pb.Component(rpg.ContainerType).(*rpg.Container).Contents() {
		fmt.Println(pb.Component(NameType), "has", o.Component(NameType))
	}

	fmt.Println("Trade succeeded:", global.Atomic(func(s *rpg.State) bool {
		pa, pb, ia, ib := s.Get(personA), s.Get(personB), s.Get(itemA), s.Get(itemB)
		if !pa.Component(rpg.ContainerType).(*rpg.Container).Remove(ia) {
			return false
		}
		if !pb.Component(rpg.ContainerType).(*rpg.Container).Remove(ib) {
			return false
		}
		if !pa.Component(rpg.ContainerType).(*rpg.Container).Add(ib) {
			return false
		}
		if !pb.Component(rpg.ContainerType).(*rpg.Container).Add(ia) {
			return false
		}
		return true
	}))

	pa = global.Get(personA)
	for _, o := range pa.Component(rpg.ContainerType).(*rpg.Container).Contents() {
		fmt.Println(pa.Component(NameType), "has", o.Component(NameType))
	}

	pb = global.Get(personB)
	for _, o := range pb.Component(rpg.ContainerType).(*rpg.Container).Contents() {
		fmt.Println(pb.Component(NameType), "has", o.Component(NameType))
	}
}
