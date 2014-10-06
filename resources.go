package rpg

type Resources struct {
	r map[string]int64
	o *Object
}

var ResourcesType = RegisterComponent(ResourcesFactory)

func ResourcesFactory(o *Object) Component {
	return &Resources{
		r: make(map[string]int64),
		o: o,
	}
}

func (r *Resources) Clone(o *Object) Component {
	clone := &Resources{
		r: make(map[string]int64, len(r.r)),
		o: o,
	}
	for id, v := range r.r {
		clone.r[id] = v
	}
	return clone
}

func (r *Resources) Get(id string) int64 {
	return r.r[id]
}

func (r *Resources) Set(id string, v int64) {
	r.r[id] = v
	r.o.Modified()
}
