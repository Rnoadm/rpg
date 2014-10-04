package rpg

type Object interface {
	ID() ObjectIndex

	Add(Object) bool
	Remove(Object) bool
	Contents() []ObjectIndex

	Clone() Object

	base() *BaseObject
}

type BaseObject struct {
	id       ObjectIndex
	contents sortedObjectIndices
	version  uint64
	modified bool
}

func (o *BaseObject) ID() ObjectIndex   { return o.id }
func (o *BaseObject) base() *BaseObject { return o }

func (o *BaseObject) Clone() Object {
	c := &BaseObject{
		id:       o.id,
		contents: make(sortedObjectIndices, len(o.contents)),
		version:  o.version,
		modified: false,
	}
	copy(c.contents, o.contents)

	return c
}

func (o *BaseObject) Add(v Object) bool {
	if o.contents.add(v.base().id) {
		o.modified = true
		return true
	}
	return false
}

func (o *BaseObject) Remove(v Object) bool {
	if o.contents.remove(v.base().id) {
		o.modified = true
		return true
	}
	return false
}

func (o *BaseObject) Contents() []ObjectIndex {
	c := make([]ObjectIndex, len(o.contents))
	copy(c, []ObjectIndex(o.contents))
	return c
}
