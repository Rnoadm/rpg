package rpg

var MaxMessages = 200

type Message struct {
	Source ObjectIndex
	Time   int64
	Text   string
	Kind   string
}

func (m *Message) String() string { return m.Text }

type Messages struct {
	m []Message
	o *Object
}

func MessagesFactory(o *Object) Component {
	return &Messages{o: o}
}

var MessagesType = RegisterComponent(MessagesFactory)

func (m *Messages) Clone(o *Object) Component {
	return &Messages{
		m: append([]Message(nil), m.m...),
		o: o,
	}
}

func (m *Messages) Len() int         { return len(m.m) }
func (m *Messages) At(i int) Message { return m.m[i] }

func (m *Messages) Append(msg Message) {
	m.m = append(m.m, msg)
	if len(m.m) > MaxMessages {
		m.m = m.m[len(m.m)-MaxMessages:]
	}
	m.o.Modified()
}
