package rpg

import "reflect"

type ComponentFactory func(*Object) Component

type Component interface {
	Clone(*Object) Component
}

type Object struct {
	id         ObjectIndex
	components map[reflect.Type]Component
	version    uint64
	modified   bool
	state      *State
}

func (o *Object) ID() ObjectIndex                    { return o.id }
func (o *Object) State() *State                      { return o.state }
func (o *Object) Component(t reflect.Type) Component { return o.components[t] }

func (o *Object) Clone(s *State) *Object {
	clone := &Object{
		id:         o.id,
		components: make(map[reflect.Type]Component, len(o.components)),
		version:    o.version,
		modified:   false,
		state:      s,
	}
	for t, c := range o.components {
		clone.components[t] = c.Clone(clone)
	}

	return clone
}

func (o *Object) Modified() {
	o.modified = true
}
