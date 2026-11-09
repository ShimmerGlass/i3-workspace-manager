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

func TestRetain(t *testing.T) {
	h := &History{Position: 2, Projects: []string{"gone", "a", "b", "gone2", "c"}}
	h.Retain([]string{"a", "b", "c"})
	if h.Position != 1 || !slices.Equal(h.Projects, []string{"a", "b", "c"}) {
		t.Fatalf("got %+v", h)
	}
	h = &History{Position: 1, Projects: []string{"a", "gone"}}
	h.Retain([]string{"a"})
	if h.Position != 0 || !slices.Equal(h.Projects, []string{"a"}) {
		t.Fatalf("got %+v", h)
	}
}

func TestAddBackground(t *testing.T) {
	tests := []struct {
		name     string
		h        History
		p        string
		wantPos  int
		wantList []string
	}{
		{"empty", History{}, "x", 0, []string{"x"}},
		{"absent", History{Position: 1, Projects: []string{"a", "b", "c"}}, "x", 2, []string{"x", "a", "b", "c"}},
		{"absent at head", History{Projects: []string{"a", "b"}}, "x", 1, []string{"x", "a", "b"}},
		{"present after current", History{Position: 1, Projects: []string{"a", "b", "c"}}, "c", 2, []string{"c", "a", "b"}},
		{"present before current", History{Position: 2, Projects: []string{"a", "b", "c"}}, "b", 2, []string{"b", "a", "c"}},
		{"present at head before current", History{Position: 1, Projects: []string{"a", "b"}}, "a", 1, []string{"a", "b"}},
		{"is current", History{Position: 2, Projects: []string{"a", "b", "c"}}, "c", 0, []string{"c", "a", "b"}},
	}
	for _, tt := range tests {
		var current string
		if tt.h.Position < len(tt.h.Projects) {
			current = tt.h.Projects[tt.h.Position]
		}
		tt.h.AddBackground(tt.p)
		if tt.h.Position != tt.wantPos || !slices.Equal(tt.h.Projects, tt.wantList) {
			t.Errorf("%s: got %+v, want %d %v", tt.name, tt.h, tt.wantPos, tt.wantList)
		}
		if current != "" && current != tt.p && tt.h.Projects[tt.h.Position] != current {
			t.Errorf("%s: current project changed to %s", tt.name, tt.h.Projects[tt.h.Position])
		}
	}
}
