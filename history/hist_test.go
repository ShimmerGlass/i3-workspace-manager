package history

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	h := &History{Position: 1, Projects: []string{"a", "b c", "d"}}
	got, err := parse(bytes.NewReader(h.format()))
	if err != nil {
		t.Fatal(err)
	}
	if got.Position != h.Position || !slices.Equal(got.Projects, h.Projects) {
		t.Fatalf("got %+v, want %+v", got, h)
	}
}

func TestFormatTruncates(t *testing.T) {
	h := &History{Position: maxSize + 10}
	for i := range maxSize + 20 {
		h.Projects = append(h.Projects, fmt.Sprint(i))
	}
	got, err := parse(bytes.NewReader(h.format()))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Projects) != maxSize || got.Position != maxSize-1 {
		t.Fatalf("got %d projects at %d", len(got.Projects), got.Position)
	}
}

func TestAdd(t *testing.T) {
	h := &History{Position: 2, Projects: []string{"a", "b", "c"}}
	h.Add("c")
	if h.Position != 0 || !slices.Equal(h.Projects, []string{"c", "a", "b"}) {
		t.Fatalf("got %+v", h)
	}
}

func TestSort(t *testing.T) {
	h := &History{Projects: []string{"c", "a"}}
	s := []string{"a", "b", "c", "d"}
	h.Sort(s)
	if want := []string{"c", "a", "b", "d"}; !slices.Equal(s, want) {
		t.Fatalf("got %v, want %v", s, want)
	}
}
