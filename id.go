package rpg

import "sort"

// ObjectIndex is the ID of an Object in a State.
type ObjectIndex uint64

type sortedObjectIndices []ObjectIndex

func (o sortedObjectIndices) Len() int           { return len(o) }
func (o sortedObjectIndices) Less(i, j int) bool { return o[i] < o[j] }
func (o sortedObjectIndices) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o *sortedObjectIndices) add(id ObjectIndex) bool {
	i := sort.Search(len(*o), func(i int) bool {
		return (*o)[i] >= id
	})

	if i < len(*o) && (*o)[i] == id {
		return false
	}

	*o = append((*o)[:i], append(sortedObjectIndices{id}, (*o)[i:]...)...)
	return true
}
func (o *sortedObjectIndices) remove(id ObjectIndex) bool {
	i := sort.Search(len(*o), func(i int) bool {
		return (*o)[i] >= id
	})

	if i < len(*o) && (*o)[i] == id {
		*o = append((*o)[:i], (*o)[i+1:]...)
		return true
	}
	return false
}
func (o *sortedObjectIndices) Push(v interface{}) {
	*o = append(*o, v.(ObjectIndex))
}
func (o *sortedObjectIndices) Pop() interface{} {
	l := len(*o) - 1
	v := (*o)[l]
	*o = (*o)[:l]
	return v
}
