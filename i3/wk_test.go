package i3

import (
	"testing"

	"github.com/joshuarubin/go-sway"
)

func view() *Node {
	pid := uint32(1)
	return &Node{Type: sway.NodeCon, PID: &pid}
}

func floatingView() *Node {
	pid := uint32(1)
	return &Node{Type: sway.NodeFloatingCon, PID: &pid}
}

func workspace(name string, nodes, floating []*Node) *Node {
	return &Node{
		Type: sway.NodeRoot,
		Nodes: []*Node{{
			Type:  sway.NodeOutput,
			Nodes: []*Node{{Type: sway.NodeWorkspace, Name: name, Nodes: nodes, FloatingNodes: floating}},
		}},
	}
}

func TestTreeWorkspaceHasWindows(t *testing.T) {
	split := func(children ...*Node) *Node { return &Node{Type: sway.NodeCon, Nodes: children} }

	tests := []struct {
		name string
		tree *Node
		want bool
	}{
		{"empty", workspace("w", nil, nil), false},
		{"direct view", workspace("w", []*Node{view()}, nil), true},
		{"view in second split", workspace("w", []*Node{split(), split(view())}, nil), true},
		{"floating container without view", workspace("w", nil, []*Node{{Type: sway.NodeFloatingCon}}), false},
		{"floating view", workspace("w", nil, []*Node{floatingView()}), true},
		{"view in floating split", workspace("w", nil, []*Node{{Type: sway.NodeFloatingCon, Nodes: []*Node{view()}}}), true},
		{"empty splits", workspace("w", []*Node{split(), split()}, nil), false},
		{"sticky floating view", workspace("w", nil, []*Node{{Type: sway.NodeFloatingCon, Sticky: true, PID: floatingView().PID}}), false},
	}
	for _, tt := range tests {
		if got := TreeWorkspaceHasWindows(tt.tree, "w"); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
	if TreeWorkspaceHasWindows(workspace("w", []*Node{view()}, nil), "other") {
		t.Error("unknown workspace reported windows")
	}
}

func TestQuote(t *testing.T) {
	if got, want := Quote(`a "b" \c`), `"a \"b\" \\c"`; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
