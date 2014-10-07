package rpg

type Location struct {
	x, y, z int64
	o       *Object
}

var LocationType = RegisterComponent(LocationFactory)

func LocationFactory(o *Object) Component {
	return &Location{o: o}
}

func (l *Location) Clone(o *Object) Component {
	return &Location{
		x: l.x,
		y: l.y,
		z: l.z,
		o: o,
	}
}

func (l *Location) Get() (x, y, z int64) {
	return l.x, l.y, l.z
}

func (l *Location) Set(x, y, z int64) {
	l.x, l.y, l.z = x, y, z
	l.o.Modified()
}

func (l *Location) Dist(o *Location) int64 {
	x := l.x - o.x
	y := l.y - o.y
	z := l.z - o.z
	return x*x + y*y + z*z
}
