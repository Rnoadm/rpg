package main

import (
	"bytes"
	"encoding/gob"
	"github.com/Rnoadm/rpg"
	"sort"
)

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
