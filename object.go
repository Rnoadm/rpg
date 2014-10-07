package rpg

import (
	"fmt"
	"reflect"
)

// ComponentFactory constructs a Component for use with the given Object. The underlying
// type of the Component must be a pointer type. The Object is nil when the function is
// called from RegisterComponent.
type ComponentFactory func(*Object) Component

// Component represents a feature of an Object.
type Component interface {
	// Clone duplicates this Component and replaces the cloned Component's Object
	// with the given Object.
	Clone(*Object) Component
}

var registeredComponents = make(map[string]ComponentFactory)

// RegisterComponent allows a ComponentFactory to be used with State.Create. It returns
// the reflect.Type of the Component which can be used in Object.Component.
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

// Object represents a person, place, or thing in a State.
type Object struct {
	id, parent ObjectIndex
	components map[reflect.Type]Component
	version    uint64
	modified   bool
	state      *State
}

// ID returns the ID of this Object.
func (o *Object) ID() ObjectIndex { return o.id }

// State returns the State this Object exists within.
func (o *Object) State() *State { return o.state }

// Component returns the Component of the given type if one exists in this Object.
func (o *Object) Component(t reflect.Type) Component { return o.components[t] }

// ComponentAny returns the Component of the given type if one exists in this Object. If
// o.Parent is non-nil, ComponentAny will try to return the parent's component, recursively.
// The returned Component should not be modified.
func (o *Object) ComponentAny(t reflect.Type) Component {
	if c, ok := o.components[t]; ok {
		return c
	}
	if p := o.Parent(); p != nil {
		return p.ComponentAny(t)
	}
	return nil
}

// ID returns the parent of this Object if it has one.
func (o *Object) Parent() *Object {
	if o.parent == 0 {
		return nil
	}
	return o.state.Get(o.parent)
}

func (o *Object) clone(s *State) *Object {
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

// Modified notifies o that one of its Components has been modified. This is required
// for State.Atomic to function properly.
func (o *Object) Modified() {
	o.modified = true
}

// Create is the same as State.Create but the Object derives from o.
func (o *Object) Create(factories ...ComponentFactory) (ObjectIndex, *Object) {
	id, o2 := o.state.Create(factories...)
	o2.parent = o.id
	return id, o2
}
