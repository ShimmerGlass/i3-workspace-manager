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

func TestTreeWorkspaceViewsKeepsSticky(t *testing.T) {
	sticky := floatingView()
	sticky.Sticky = true
	tree := workspace("w", []*Node{view()}, []*Node{sticky})
	if got := len(TreeWorkspaceViews(tree, "w")); got != 2 {
		t.Errorf("got %d views, want 2", got)
	}
}

func TestParseScratchpadIDs(t *testing.T) {
	data := []byte(`{"id":1,"scratchpad_state":"none","nodes":[{"id":2,"nodes":[{"id":3,"scratchpad_state":"none"}],"floating_nodes":[{"id":4,"scratchpad_state":"fresh"},{"id":5,"scratchpad_state":"changed"},{"id":6,"scratchpad_state":"none"}]}]}`)
	ids, err := parseScratchpadIDs(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || !ids[4] || !ids[5] {
		t.Errorf("got %v, want 4 and 5", ids)
	}
	if _, err := parseScratchpadIDs([]byte("nope")); err == nil {
		t.Error("bad json accepted")
	}
}

func TestQuote(t *testing.T) {
	if got, want := Quote(`a "b" \c`), `"a \"b\" \\c"`; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestContainerWorkspace(t *testing.T) {
	v := view()
	v.ID = 7
	f := floatingView()
	f.ID = 8
	tree := workspace("w", []*Node{{Type: sway.NodeCon, Nodes: []*Node{v}}}, []*Node{f})
	tree.Nodes[0].Nodes = append(tree.Nodes[0].Nodes, &Node{Type: sway.NodeWorkspace, Name: "other", Nodes: []*Node{func() *Node { n := view(); n.ID = 9; return n }()}})

	for id, want := range map[int64]string{7: "w", 8: "w", 9: "other"} {
		if got, ok := ContainerWorkspace(tree, id); !ok || got != want {
			t.Errorf("id %d: got %q %v, want %q", id, got, ok, want)
		}
	}
	if _, ok := ContainerWorkspace(tree, 99); ok {
		t.Error("unknown container found")
	}
}
