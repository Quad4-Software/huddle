package hub

import "testing"

func TestClientSlicePoolReuse(t *testing.T) {
	ptr, s := borrowClientSlice(4)
	s = append(s, &Client{ID: "a"}, &Client{ID: "b"})
	*ptr = s
	if len(s) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(s))
	}
	releaseClientSlice(ptr)

	ptr2, s2 := borrowClientSlice(4)
	if cap(s2) < 2 {
		t.Fatalf("expected pooled slice capacity, got %d", cap(s2))
	}
	releaseClientSlice(ptr2)
}
