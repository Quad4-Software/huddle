package hub

import "sync"

const defaultClientSliceCap = 16

var clientSlicePool = sync.Pool{
	New: func() any {
		s := make([]*Client, 0, defaultClientSliceCap)
		return &s
	},
}

func borrowClientSlice(capacity int) (*[]*Client, []*Client) {
	ptr := clientSlicePool.Get().(*[]*Client)
	s := (*ptr)[:0]
	if cap(s) < capacity {
		s = make([]*Client, 0, capacity)
	}
	*ptr = s
	return ptr, s
}

func releaseClientSlice(ptr *[]*Client) {
	s := *ptr
	if cap(s) > 256 {
		return
	}
	*ptr = s[:0]
	clientSlicePool.Put(ptr)
}
