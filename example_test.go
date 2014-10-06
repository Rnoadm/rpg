package rpg_test

import (
	"fmt"
	"github.com/Rnoadm/rpg"
	"github.com/Rnoadm/rpg/history"
	"io"
	"io/ioutil"
	"os"
)

type Name string

func NameFactory(name string) rpg.ComponentFactory {
	n := Name(name)
	return n.Clone
}

var NameType = rpg.RegisterComponent(NameFactory(""))

func (n *Name) Clone(*rpg.Object) rpg.Component {
	clone := *n
	return &clone
}

func (n *Name) String() string {
	return string(*n)
}

func printContainers(s *rpg.State, ids ...rpg.ObjectIndex) {
	for _, id := range ids {
		p := s.Get(id)
		for _, o := range p.Component(rpg.ContainerType).(*rpg.Container).Contents() {
			fmt.Println(p.Component(NameType), "has", o.Component(NameType))
		}
	}
}

func Example() {
	f, err := ioutil.TempFile(os.TempDir(), "rnoadm")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	h := history.NewHistory(f)

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
	err = h.Append(global)
	if err != nil {
		panic(err)
	}

	printContainers(global, personA, personB)

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

	err = h.Append(global)
	if err != nil {
		panic(err)
	}

	printContainers(global, personA, personB)

	fmt.Println()
	h.Reset()
	for {
		s, err := h.Seek(1, history.SeekCur)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		fmt.Println("Forward:", h.Tell())

		printContainers(s, personA, personB)
	}

	fmt.Println()
	h.Reset()
	for {
		s, err := h.Seek(-1, history.SeekCur)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		fmt.Println("Reverse:", h.Tell())

		printContainers(s, personA, personB)
	}
	// Output:
	// person A has item A
	// person B has item B
	// Trade succeeded: true
	// person A has item B
	// person B has item A
	//
	// Forward: 0
	// person A has item A
	// person B has item B
	// Forward: 1
	// person A has item B
	// person B has item A
	//
	// Reverse: 1
	// person A has item B
	// person B has item A
	// Reverse: 0
	// person A has item A
	// person B has item B
}
