// +build ignore

package main

import (
	"fmt"
	"github.com/Rnoadm/rpg"
)

type Person struct {
	rpg.Object

	name string
}

func PersonFactory(base rpg.Object) rpg.Object {
	return &Person{Object: base}
}

func (p *Person) String() string {
	return p.name
}

func (p *Person) Clone() rpg.Object {
	return &Person{
		Object: p.Object.Clone(),

		name: p.name,
	}
}

type Item struct {
	rpg.Object

	name string
}

func ItemFactory(base rpg.Object) rpg.Object {
	return &Item{Object: base}
}

func (i *Item) String() string {
	return i.name
}

func (i *Item) Clone() rpg.Object {
	return &Item{
		Object: i.Object.Clone(),

		name: i.name,
	}
}

func main() {
	global := rpg.NewState()

	var personA, personB, itemA, itemB rpg.ObjectIndex
	if !global.Atomic(func(s *rpg.State) bool {
		var pa, pb, ia, ib rpg.Object
		personA, pa = s.Create(PersonFactory)
		personB, pb = s.Create(PersonFactory)
		itemA, ia = s.Create(ItemFactory)
		itemB, ib = s.Create(ItemFactory)

		pa.(*Person).name = "Person A"
		pb.(*Person).name = "Person B"
		ia.(*Item).name = "Item A"
		ib.(*Item).name = "Item B"

		if !pa.Add(ia) {
			panic("unreachable")
		}
		if !pb.Add(ib) {
			panic("unreachable")
		}
		return true
	}) {
		panic("unreachable")
	}

	pa := global.Get(personA)
	for _, id := range pa.Contents() {
		fmt.Println(pa, "has", global.Get(id))
	}

	pb := global.Get(personB)
	for _, id := range pb.Contents() {
		fmt.Println(pb, "has", global.Get(id))
	}

	fmt.Println("Trade succeeded:", global.Atomic(func(s *rpg.State) bool {
		pa, pb, ia, ib := s.Get(personA), s.Get(personB), s.Get(itemA), s.Get(itemB)
		if !pa.Remove(ia) {
			return false
		}
		if !pb.Remove(ib) {
			return false
		}
		if !pa.Add(ib) {
			return false
		}
		if !pb.Add(ia) {
			return false
		}
		return true
	}))

	pa = global.Get(personA)
	for _, id := range pa.Contents() {
		fmt.Println(pa, "has", global.Get(id))
	}

	pb = global.Get(personB)
	for _, id := range pb.Contents() {
		fmt.Println(pb, "has", global.Get(id))
	}
}
