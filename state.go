package rpg

import (
	"sync"
	"sync/atomic"
)

type ObjectFactory func(Object) Object

type State struct {
	parent  *State
	objects map[ObjectIndex]Object
	mtx     sync.Mutex
}

func NewState() *State {
	return &State{objects: make(map[ObjectIndex]Object)}
}

func (s *State) Atomic(f func(*State) bool) bool {
retry:
	for {
		child := NewState()
		child.parent = s
		if !f(child) {
			return false
		}

		child.mtx.Lock()
		defer child.mtx.Unlock()

		s.mtx.Lock()
		defer s.mtx.Unlock()

		for id, o := range child.objects {
			if !o.base().modified {
				continue
			}
			if p, ok := s.objects[id]; ok && p.base().version != o.base().version {
				continue retry
			}
		}

		for id, o := range child.objects {
			if !o.base().modified {
				continue
			}
			o.base().version = atomic.AddUint64(&nextObjectVersion, 1)
			s.objects[id] = o
		}
		return true
	}
}

func (s *State) Create(f ObjectFactory) (id ObjectIndex, o Object) {
	id = ObjectIndex(atomic.AddUint64(&nextObjectID, 1))
	o = f(&BaseObject{
		id:       id,
		version:  atomic.AddUint64(&nextObjectVersion, 1),
		modified: true,
	})

	s.mtx.Lock()
	s.objects[id] = o
	s.mtx.Unlock()

	return
}

func (s *State) Get(id ObjectIndex) Object {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if o, ok := s.objects[id]; ok {
		return o
	}

	o := s.parent.Get(id).Clone()
	s.objects[id] = o

	return o
}
