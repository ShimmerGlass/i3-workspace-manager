package main

import (
	"slices"
	"testing"
)

func TestParseLines(t *testing.T) {
	got := parseLines("b\n\n  a  \nb\r\n")
	if want := []string{"a", "b"}; !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}
