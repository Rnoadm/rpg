package rpg_test

import (
	"fmt"
	"github.com/Rnoadm/rpg"
	"github.com/Rnoadm/rpg/history"
	"io"
	"io/ioutil"
	"os"
)

func printContainers(s *rpg.State) {
	for _, id := range s.IDs() {
		p := s.Get(id)
		if container, ok := p.Component(rpg.ContainerType).(*rpg.Container); ok {
			for _, o := range container.Contents() {
				fmt.Println(p.Component(rpg.NameType), "has", o.Component(rpg.NameType))
			}
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
		personA, pa = s.Create(rpg.NameFactory("person A"), rpg.ContainerFactory)
		personB, pb = s.Create(rpg.NameFactory("person B"), rpg.ContainerFactory)
		itemA, ia = s.Create(rpg.NameFactory("item A"))
		itemB, ib = s.Create(rpg.NameFactory("item B"))

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

	printContainers(global)

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

	printContainers(global)

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
		printContainers(s)
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
		printContainers(s)
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
