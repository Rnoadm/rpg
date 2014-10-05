package rpg

import (
	"fmt"
	"reflect"
)

type ComponentFactory func(*Object) Component

type Component interface {
	Clone(*Object) Component
}

var (
	registeredComponents = make(map[string]ComponentFactory)
)

func RegisterComponent(f ComponentFactory) reflect.Type {
	c := f(nil)
	t := reflect.TypeOf(c)
	registeredComponents[typeName(t)] = f
	return t
}

func typeName(t reflect.Type) string {
	// copied with modifications from encoding/gob
	star := ""
	if t.Name() == "" && t.Kind() == reflect.Ptr {
		star = "*"
		t = t.Elem()
	}
	if t.Name() == "" {
		return t.String()
	}
	return fmt.Sprintf("%s%q.%s", star, t.PkgPath(), t.Name())
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
