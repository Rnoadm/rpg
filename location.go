package rpg

// Location represents the position of an Object.
type Location struct {
	x, y, z int64
	o       *Object
}

// LocationFactory is a ComponentFactory.
func LocationFactory(o *Object) Component {
	return &Location{o: o}
}

// LocationType can be used with Object.Component to retrieve a Location.
var LocationType = RegisterComponent(LocationFactory)

// Clone implements Component.
func (l *Location) Clone(o *Object) Component {
	return &Location{
		x: l.x,
		y: l.y,
		z: l.z,
		o: o,
	}
}

// Get returns the position represented by l.
func (l *Location) Get() (x, y, z int64) {
	return l.x, l.y, l.z
}

// Set modifies l, along with any Location-having Objects inside a Container.
func (l *Location) Set(x, y, z int64) {
	l.x, l.y, l.z = x, y, z
	if c, ok := l.o.Component(ContainerType).(*Container); ok {
		for _, o := range c.Contents() {
			if l2, ok := o.Component(LocationType).(*Location); ok {
				l2.Set(x, y, z)
			}
		}
	}
	l.o.Modified()
}

// Dist returns the distance squared between l and o.
func (l *Location) Dist(o *Location) int64 {
	x := l.x - o.x
	y := l.y - o.y
	z := l.z - o.z
	return x*x + y*y + z*z
}

// Equal returns true if l and o have the same return values for Get.
func (l *Location) Equal(o *Location) bool {
	return l.x == o.x && l.y == o.y && l.z == o.z
}
