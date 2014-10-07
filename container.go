package rpg

// Container is a Component that holds references to other Objects.
type Container struct {
	c sortedObjectIndices
	o *Object
}

// ContainerFactory is a ComponentFactory.
func ContainerFactory(o *Object) Component {
	return &Container{o: o}
}

// ContainerType can be used with Object.Component to retrieve a Container.
var ContainerType = RegisterComponent(ContainerFactory)

// Clone implements Component.
func (c *Container) Clone(o *Object) Component {
	clone := &Container{
		c: make(sortedObjectIndices, len(c.c)),
		o: o,
	}
	copy(clone.c, c.c)
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
		c.o.Modified()
		return true
	}
	return false
}

// Remove removes v from c, returning false if v is not in c.
func (c *Container) Remove(v *Object) bool {
	if c.c.remove(v.ID()) {
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
