package rpg

type Container struct {
	c sortedObjectIndices
	o *Object
}

func ContainerFactory(o *Object) Component {
	return &Container{o: o}
}

var ContainerType = RegisterComponent(ContainerFactory)

func (c *Container) Clone(o *Object) Component {
	clone := &Container{
		c: make(sortedObjectIndices, len(c.c)),
		o: o,
	}
	copy(clone.c, c.c)
	return clone
}

func (c *Container) Add(v *Object) bool {
	if c.c.add(v.ID()) {
		c.o.Modified()
		return true
	}
	return false
}

func (c *Container) Remove(v *Object) bool {
	if c.c.remove(v.ID()) {
		c.o.Modified()
		return true
	}
	return false
}

func (c *Container) Contents() []*Object {
	contents := make([]*Object, len(c.c))
	for i, id := range c.c {
		contents[i] = c.o.State().Get(id)
	}
	return contents
}
