package main

import (
	"slices"
	"testing"

	"github.com/shimmerglass/i3-workspace-manager/i3"
)

func TestParseLines(t *testing.T) {
	got := parseLines("b\n\n  a  \nb\r\n")
	if want := []string{"a", "b"}; !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestProjectWindow(t *testing.T) {
	tree := &i3.Node{ID: 1, Nodes: []*i3.Node{
		{ID: 2, Name: "notes.txt — other"},
		{ID: 3, Name: "argocd-no-sync — main.go"},
	}}
	if id, ok := projectWindow(tree, []int64{2, 3}, "VoodooTeam/argocd-casual.t/argocd-no-sync"); !ok || id != 3 {
		t.Fatalf("got %d %v, want 3 true", id, ok)
	}
	if _, ok := projectWindow(tree, []int64{2}, "VoodooTeam/argocd-casual.t/argocd-no-sync"); ok {
		t.Fatal("unrelated window matched")
	}
	if _, ok := projectWindow(tree, []int64{42}, "x"); ok {
		t.Fatal("unknown id matched")
	}
}
