package rpg

import "reflect"

// Container is a Component that holds references to other Objects.
type Container struct {
	c sortedObjectIndices
	o *Object

	by_component map[reflect.Type]sortedObjectIndices
}

// ContainerFactory is a ComponentFactory.
func ContainerFactory(o *Object) Component {
	return &Container{o: o, by_component: make(map[reflect.Type]sortedObjectIndices)}
}

// ContainerType can be used with Object.Component to retrieve a Container.
var ContainerType = RegisterComponent(ContainerFactory)

// Clone implements Component.
func (c *Container) Clone(o *Object) Component {
	c.needComponents()
	clone := &Container{
		c:            append(sortedObjectIndices(nil), c.c...),
		o:            o,
		by_component: make(map[reflect.Type]sortedObjectIndices, len(c.by_component)),
	}
	for t, b := range c.by_component {
		clone.by_component[t] = append(sortedObjectIndices(nil), b...)
	}
	return clone
}

// Add adds v to c, returning false if v is already in c. If v has a Location, it is modified
// to the containing object's Location.
func (c *Container) Add(v *Object) bool {
	if c.c.add(v.ID()) {
		if l1, ok := c.o.Component(LocationType).(*Location); ok {
			if l2, ok := v.Component(LocationType).(*Location); ok {
				l2.Set(l1.Get())
			}
		}
		c.needComponents()
		for t := range v.components {
			b := c.by_component[t]
			b.add(v.ID())
			c.by_component[t] = b
		}
		c.o.Modified()
		return true
	}
	return false
}

// Remove removes v from c, returning false if v is not in c.
func (c *Container) Remove(v *Object) bool {
	if c.c.remove(v.ID()) {
		c.needComponents()
		for t := range v.components {
			b := c.by_component[t]
			b.remove(v.ID())
			c.by_component[t] = b
		}
		c.o.Modified()
		return true
	}
	return false
}

// Contents returns the sorted set of Objects in c.
func (c *Container) Contents() []*Object {
	contents := make([]*Object, len(c.c))
	for i, id := range c.c {
		contents[i] = c.o.State().Get(id)
	}
	return contents
}

// ByComponent returns the sorted set of Objects that have Component t in c.
func (c *Container) ByComponent(t reflect.Type) []*Object {
	c.needComponents()
	b := c.by_component[t]
	contents := make([]*Object, len(b))
	for i, id := range b {
		contents[i] = c.o.State().Get(id)
	}
	return contents
}

func (c *Container) needComponents() {
	if c.by_component != nil {
		return
	}

	c.by_component = make(map[reflect.Type]sortedObjectIndices)
	for _, o := range c.Contents() {
		for t := range o.components {
			b := c.by_component[t]
			b.add(o.ID())
			c.by_component[t] = b
		}
	}
}
