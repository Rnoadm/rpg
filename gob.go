package rpg

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
	"reflect"
	"sync/atomic"
)

func readUvarint(buf []byte) (uint64, []byte, error) {
	x, i := binary.Uvarint(buf)
	if i == 0 {
		return x, buf, io.ErrUnexpectedEOF
	}
	return x, buf[i:], nil
}

func writeUvarint(buf []byte, x uint64) []byte {
	var b [binary.MaxVarintLen64]byte
	i := binary.PutUvarint(b[:], x)
	return append(buf, b[:i]...)
}

func readString(buf []byte) (string, []byte, error) {
	l, buf, err := readUvarint(buf)
	if err != nil {
		return "", buf, err
	}
	if l > uint64(len(buf)) {
		return "", buf, io.ErrUnexpectedEOF
	}
	return string(buf[:l]), buf[l:], nil
}

func writeString(buf []byte, s string) []byte {
	buf = writeUvarint(buf, uint64(len(s)))
	return append(buf, s...)
}

var (
	ErrEncodeStateParent         = errors.New("rpg: cannot encode a child State")
	ErrEncodeStateVersion        = errors.New("rpg: unrecognized State version")
	ErrEncodeObjectStateless     = errors.New("rpg: cannot decode Object directly")
	ErrEncodeObjectVersion       = errors.New("rpg: unrecognized Object version")
	ErrEncodeContainerVersion    = errors.New("rpg: unrecognized Container version")
	ErrEncodeContainerOutOfOrder = errors.New("rpg: Container is out of order")
)

const (
	stateVersion     = 0
	objectVersion    = 0
	containerVersion = 0
)

func (s *State) GobEncode() (data []byte, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.parent != nil {
		return nil, ErrEncodeStateParent
	}

	data = writeUvarint(data, stateVersion)
	data = writeUvarint(data, atomic.LoadUint64(s.nextObjectID))
	data = writeUvarint(data, uint64(len(s.objects)))
	h := make(sortedObjectIndices, 0, len(s.objects))
	for id := range s.objects {
		heap.Push(&h, id)
	}
	objects := make([]*Object, 0, len(s.objects))
	for len(h) > 0 {
		id := heap.Pop(&h).(ObjectIndex)
		data = writeUvarint(data, uint64(id))
		objects = append(objects, s.objects[id])
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	for _, o := range objects {
		err = enc.Encode(o)
		if err != nil {
			return
		}
	}
	data = append(data, buf.Bytes()...)
	return
}

func (s *State) GobDecode(data []byte) (err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.parent != nil {
		return ErrEncodeStateParent
	}

	version, data, err := readUvarint(data)
	if err != nil {
		return
	}
	if version != stateVersion {
		return ErrEncodeStateVersion
	}
	if s.nextObjectID == nil {
		s.objects = make(map[ObjectIndex]*Object)
		s.nextObjectID = new(uint64)
		s.nextObjectVersion = new(uint64)
	}

	nextID, data, err := readUvarint(data)
	if err != nil {
		return
	}
	atomic.StoreUint64(s.nextObjectID, nextID)

	objectCount, data, err := readUvarint(data)
	if err != nil {
		return
	}

	objects := make([]*Object, objectCount)
	for i := range objects {
		var id uint64
		id, data, err = readUvarint(data)
		if err != nil {
			return
		}
		objects[i] = &Object{id: ObjectIndex(id), state: s}
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	s.objects = make(map[ObjectIndex]*Object, objectCount)
	for _, o := range objects {
		err = dec.Decode(o)
		if err != nil {
			return
		}
		s.objects[o.id] = o
	}
	return
}

type componentHeapElement struct {
	t string
	c Component
}
type componentHeap []componentHeapElement

func (h *componentHeap) Len() int           { return len(*h) }
func (h *componentHeap) Swap(i, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }
func (h *componentHeap) Less(i, j int) bool { return (*h)[i].t < (*h)[j].t }
func (h *componentHeap) Push(v interface{}) {
	*h = append(*h, v.(componentHeapElement))
}
func (h *componentHeap) Pop() interface{} {
	l := len(*h) - 1
	v := (*h)[l]
	*h = (*h)[:l]
	return v
}

func (o *Object) GobEncode() (data []byte, err error) {
	data = writeUvarint(data, objectVersion)

	h := make(componentHeap, 0, len(o.components))
	for t, c := range o.components {
		heap.Push(&h, componentHeapElement{t: typeName(t), c: c})
	}

	data = writeUvarint(data, uint64(len(o.components)))
	components := make([]Component, 0, len(o.components))
	for len(h) != 0 {
		c := heap.Pop(&h).(componentHeapElement)
		data = writeString(data, c.t)
		components = append(components, c.c)
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	for _, c := range components {
		err = enc.Encode(c)
		if err != nil {
			return
		}
	}
	data = append(data, buf.Bytes()...)

	return
}

func (o *Object) GobDecode(data []byte) (err error) {
	if o.state == nil {
		return ErrEncodeObjectStateless
	}

	version, data, err := readUvarint(data)
	if err != nil {
		return
	}
	if version != objectVersion {
		return ErrEncodeObjectVersion
	}

	componentCount, data, err := readUvarint(data)
	if err != nil {
		return
	}

	components := make([]Component, componentCount)
	for i := range components {
		var tn string
		tn, data, err = readString(data)
		if err != nil {
			return
		}
		f, ok := registeredComponents[tn]
		if !ok {
			return errors.New("rpg: unregistered component factory " + tn)
		}
		components[i] = f(o)
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	o.components = make(map[reflect.Type]Component, componentCount)
	for _, c := range components {
		err = dec.Decode(c)
		if err != nil {
			return
		}
		t := reflect.TypeOf(c)
		if _, ok := o.components[t]; ok {
			panic("rpg: multiple components of type " + t.String())
		}
		o.components[t] = c
	}
	return
}

func (c *Container) GobEncode() (data []byte, err error) {
	data = writeUvarint(data, containerVersion)
	data = writeUvarint(data, uint64(len(c.c)))
	for _, id := range c.c {
		data = writeUvarint(data, uint64(id))
	}
	return
}

func (c *Container) GobDecode(data []byte) (err error) {
	version, data, err := readUvarint(data)
	if err != nil {
		return
	}
	if version != containerVersion {
		return ErrEncodeContainerVersion
	}
	count, data, err := readUvarint(data)
	if err != nil {
		return
	}
	c.c = make(sortedObjectIndices, count)
	for i := range c.c {
		var id uint64
		id, data, err = readUvarint(data)
		if err != nil {
			return
		}
		c.c[i] = ObjectIndex(id)
	}
	for i := len(c.c) - 1; i > 0; i-- {
		if c.c[i] < c.c[i-1] {
			return ErrEncodeContainerOutOfOrder
		}
	}
	return
}
