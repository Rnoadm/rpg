package rpg

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type State struct {
	parent  *State
	objects map[ObjectIndex]*Object
	mtx     sync.Mutex

	nextObjectID, nextObjectVersion *uint64
}

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

func (s *State) Get(id ObjectIndex) *Object {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if o, ok := s.objects[id]; ok {
		return o
	}

	o := s.parent.Get(id).Clone(s)
	s.objects[id] = o

	return o
}
