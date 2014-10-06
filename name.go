package rpg

type Name string

func NameFactory(name string) ComponentFactory {
	n := Name(name)
	return n.Clone
}

var NameType = RegisterComponent(NameFactory(""))

func (n *Name) Clone(*Object) Component {
	clone := *n
	return &clone
}

func (n *Name) String() string {
	return string(*n)
}
