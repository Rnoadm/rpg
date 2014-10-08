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
	parent       *State
	objects      map[ObjectIndex]*Object
	by_component map[reflect.Type][]ObjectIndex
	mtx          sync.Mutex

	nextObjectID, nextObjectVersion *uint64

	deleted        map[ObjectIndex]uint64
	deletedVersion uint64
}

// NewState initializes an empty State.
func NewState() *State {
	return newState(nil)
}

func newState(parent *State) *State {
	s := &State{
		parent:       parent,
		objects:      make(map[ObjectIndex]*Object),
		by_component: make(map[reflect.Type][]ObjectIndex),
		deleted:      make(map[ObjectIndex]uint64),
	}
	if parent == nil {
		var a [2]uint64
		s.nextObjectID, s.nextObjectVersion = &a[0], &a[1]
	} else {
		s.nextObjectID, s.nextObjectVersion = parent.nextObjectID, parent.nextObjectVersion
		parent.mtx.Lock()
		defer parent.mtx.Unlock()

		s.deletedVersion = parent.deletedVersion
		for id, v := range parent.deleted {
			s.deleted[id] = v
		}
	}
	return s
}

// Atomic calls f and tries to apply its changes. This is the only way a State should be
// modified. f may be called multiple times if other calls to Atomic are being processed
// at the same time. Returning false from f causes Atomic to return false without
// applying the changes.
//
// This does not currently work recursively, but it may in the future.
func (s *State) Atomic(f func(*State) bool) bool {
	// TODO: make this work recursively

	for {
		child := newState(s)
		if !f(child) {
			return false
		}

		if func() bool {
			child.mtx.Lock()
			defer child.mtx.Unlock()

			s.mtx.Lock()
			defer s.mtx.Unlock()

			if child.deletedVersion != s.deletedVersion {
				for id, v1 := range s.deleted {
					if v2, ok := child.deleted[id]; !ok || v1 != v2 {
						return false
					}
				}
			}

			var newlyDeleted []ObjectIndex

			for id, o := range child.objects {
				if o == nil {
					if p, ok := s.objects[id]; ok && p != nil {
						if child.deleted[id] != p.version {
							return false
						}
						newlyDeleted = append(newlyDeleted, id)
					}
					continue
				}
				if !o.modified {
					continue
				}
				if p, ok := s.objects[id]; ok && p.version != o.version {
					return false
				}
			}

			for id, o := range child.objects {
				if o == nil {
					s.objects[id] = nil
					continue
				}
				if !o.modified {
					continue
				}
				o.version = atomic.AddUint64(s.nextObjectVersion, 1)
				s.objects[id] = o
			}
			for t, m := range child.by_component {
				s.by_component[t] = append(s.by_component[t], m...)
			}
			if len(newlyDeleted) != 0 {
				for t, m := range s.by_component {
					ids := sortedObjectIndices(m)
					sort.Sort(ids)
					removed := false
					for _, id := range newlyDeleted {
						if ids.remove(id) {
							removed = true
						}
					}
					if removed {
						s.by_component[t] = []ObjectIndex(ids)
					}
				}
			}
			s.deleted = child.deleted
			s.deletedVersion = child.deletedVersion
			return true
		}() {
			return true
		}
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
	for t := range o.components {
		s.by_component[t] = append(s.by_component[t], id)
	}
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

// Delete removes an object from the State. Future calls to Get will return nil. If the
// object is referenced anywhere, Bad Things™ will happen.
func (s *State) Delete(id ObjectIndex) {
	o := s.Get(id)
	if o == nil {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.objects[id] = nil
	s.deleted[id] = o.version
	s.deletedVersion = atomic.AddUint64(s.nextObjectVersion, 1)
}

// IDs returns the set of ObjectIndex accessible from s in ascending order.
func (s *State) IDs() []ObjectIndex {
	ids, remove := s.appendIDs(nil, nil)
	sort.Sort(ids)
	for _, r := range remove {
		ids.remove(r)
	}
	return []ObjectIndex(ids)
}

func (s *State) appendIDs(ids sortedObjectIndices, remove []ObjectIndex) (sortedObjectIndices, []ObjectIndex) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.parent != nil {
		ids, remove = s.parent.appendIDs(ids, remove)
	}

	for id, o := range s.objects {
		if s.parent == nil || !s.parent.hasID(id) {
			ids = append(ids, id)
		}
		if o == nil {
			remove = append(remove, id)
		}
	}

	return ids, remove
}

func (s *State) hasID(id ObjectIndex) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if o, ok := s.objects[id]; ok {
		return o != nil
	}
	if s.parent == nil {
		return false
	}
	return s.parent.hasID(id)
}

// ByComponent returns a sorted set of IDs of objects that have the given component type.
func (s *State) ByComponent(t reflect.Type) []ObjectIndex {
	ids := s.byComponent(t)
	sort.Sort(sortedObjectIndices(ids))
	return ids
}

func (s *State) byComponent(t reflect.Type) []ObjectIndex {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	var ids []ObjectIndex
	if s.parent != nil {
		ids = s.parent.byComponent(t)
	}

	return append(ids, s.by_component[t]...)
}

func (s *State) clearDeleted() {
	for id := range s.deleted {
		delete(s.objects, id)
	}
	s.deleted = make(map[ObjectIndex]uint64)
}
