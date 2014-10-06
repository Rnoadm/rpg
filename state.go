// Package rpg provides a base API for a role playing game.
package rpg

import (
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
)

// State represents a set of Object that can be modified concurrently using compare-and-set.
type State struct {
	parent  *State
	objects map[ObjectIndex]*Object
	mtx     sync.Mutex

	nextObjectID, nextObjectVersion *uint64
}

// NewState initializes an empty State.
func NewState() *State {
	return newState(nil)
}

func newState(parent *State) *State {
	s := &State{
		parent:  parent,
		objects: make(map[ObjectIndex]*Object),
	}
	if parent == nil {
		var a [2]uint64
		s.nextObjectID, s.nextObjectVersion = &a[0], &a[1]
	} else {
		s.nextObjectID, s.nextObjectVersion = parent.nextObjectID, parent.nextObjectVersion
	}
	return s
}

// Atomic calls f and tries to apply its changes. This is the only way a State should be
// modified. f may be called multiple times if other calls to Atomic are being processed
// at the same time. Returning false from f causes Atomic to return false without
// applying the changes.
func (s *State) Atomic(f func(*State) bool) bool {
retry:
	for {
		child := newState(s)
		if !f(child) {
			return false
		}

		child.mtx.Lock()
		defer child.mtx.Unlock()

		s.mtx.Lock()
		defer s.mtx.Unlock()

		for id, o := range child.objects {
			if !o.modified {
				continue
			}
			if p, ok := s.objects[id]; ok && p.version != o.version {
				continue retry
			}
		}

		for id, o := range child.objects {
			if !o.modified {
				continue
			}
			o.version = atomic.AddUint64(s.nextObjectVersion, 1)
			s.objects[id] = o
		}
		return true
	}
}

// Create initializes a new Object and returns it and its ObjectIndex. The factories
// must not be duplicate and must be pre-registered. The id is unique for all Objects
// in this State heirarchy.
func (s *State) Create(factories ...ComponentFactory) (id ObjectIndex, o *Object) {
	id = ObjectIndex(atomic.AddUint64(s.nextObjectID, 1))
	o = &Object{
		id:         id,
		components: make(map[reflect.Type]Component, len(factories)),
		version:    atomic.AddUint64(s.nextObjectVersion, 1),
		modified:   true,
		state:      s,
	}

	for _, f := range factories {
		c := f(o)
		t := reflect.TypeOf(c)
		if _, ok := registeredComponents[typeName(t)]; !ok {
			panic("rpg: unregistered component type " + t.String())
		}
		if _, ok := o.components[t]; ok {
			panic("rpg: multiple components of type " + t.String())
		}
		o.components[t] = c
	}

	s.mtx.Lock()
	s.objects[id] = o
	s.mtx.Unlock()

	return
}

// Get returns the Object identified by id. The object is specific to this State.
func (s *State) Get(id ObjectIndex) *Object {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if o, ok := s.objects[id]; ok {
		return o
	}
	if s.parent == nil {
		return nil
	}

	o := s.parent.Get(id)
	if o == nil {
		return nil
	}

	o = o.clone(s)
	s.objects[id] = o

	return o
}

// IDs returns the set of ObjectIndex accessible from s in ascending order.
func (s *State) IDs() []ObjectIndex {
	ids := s.appendIDs(nil)
	sort.Sort(sortedObjectIndices(ids))
	return ids
}

func (s *State) appendIDs(ids []ObjectIndex) []ObjectIndex {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.parent != nil {
		ids = s.parent.appendIDs(ids)
	}

	for id := range s.objects {
		if s.parent == nil || !s.parent.hasID(id) {
			ids = append(ids, id)
		}
	}

	return ids
}

func (s *State) hasID(id ObjectIndex) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.objects[id]; ok {
		return true
	}
	if s.parent == nil {
		return false
	}
	return s.parent.hasID(id)
}
