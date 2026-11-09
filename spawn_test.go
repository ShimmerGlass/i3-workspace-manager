package main

import (
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/shimmerglass/i3-workspace-manager/i3"
)

var nastyStrings = []string{
	"plain",
	"with space",
	"it's",
	`say "hi"`,
	"a;b",
	"a,b",
	`back\slash`,
	`$HOME and $(echo x)`,
	`'; "; \ , $ `,
	"  leading and trailing  ",
	"uni​code",
	"",
}

func TestShellQuote(t *testing.T) {
	for _, s := range nastyStrings {
		out, err := exec.Command("/bin/sh", "-c", "printf %s "+shellQuote(s)).Output()
		if err != nil {
			t.Fatalf("%q: %v", s, err)
		}
		if string(out) != s {
			t.Errorf("got %q, want %q", out, s)
		}
	}
}

func TestBuildSpawnCommandRoundTrip(t *testing.T) {
	wrapper := regexp.MustCompile(`^exec /bin/sh -c 'eval "\$\(printf %s [A-Za-z0-9+/=]+ \| base64 -d\)"'$`)
	for _, project := range nastyStrings {
		for _, command := range nastyStrings {
			if command == "" {
				continue
			}
			sc := `printf '%s\n%s\n' "$PROJECT_NAME" ` + shellQuote(command)
			got, err := buildSpawnCommand("20:w", "", project, sc, nil)
			if err != nil {
				t.Fatal(err)
			}
			parts := strings.Split(got, "; ")
			if len(parts) != 2 {
				t.Fatalf("project %q command %q: sway would split into %d commands: %s", project, command, len(parts), got)
			}
			if !wrapper.MatchString(parts[1]) {
				t.Fatalf("unsafe sway text: %s", parts[1])
			}
			execCmd := strings.TrimPrefix(parts[1], "exec ")
			out, err := exec.Command("/bin/sh", "-c", execCmd).Output()
			if err != nil {
				t.Fatalf("project %q command %q: %v", project, command, err)
			}
			if want := project + "\n" + command + "\n"; string(out) != want {
				t.Errorf("got %q, want %q", out, want)
			}
		}
	}
}

func TestBuildSpawnCommandShape(t *testing.T) {
	got, err := buildSpawnCommand("20:w", "DP-1", "p", "true", []string{"1", `we"ird`})
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(got, "; ")
	if len(parts) != 5 {
		t.Fatalf("got %d parts: %s", len(parts), got)
	}
	want := []string{
		`workspace "20:w" output "DP-1"`,
		`workspace --no-auto-back-and-forth "20:w"`,
	}
	if !slices.Equal(parts[:2], want) {
		t.Errorf("got %q, want %q", parts[:2], want)
	}
	if !strings.HasPrefix(parts[2], "exec ") {
		t.Errorf("third part %q is not an exec", parts[2])
	}
	if parts[3] != `workspace --no-auto-back-and-forth "1"` || parts[4] != `workspace --no-auto-back-and-forth "we\"ird"` {
		t.Errorf("bad restore: %q", parts[3:])
	}

	got, err = buildSpawnCommand("20:w", "", "p", "true", nil)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(got, "; "); n != 1 {
		t.Errorf("foreground without assign has %d separators: %s", n, got)
	}
}

func TestBuildSpawnCommandRejectsNewlines(t *testing.T) {
	if _, err := buildSpawnCommand("a\nb", "", "p", "true", nil); err == nil {
		t.Error("newline in workspace name accepted")
	}
	if _, err := buildSpawnCommand("a", "", "p", "true", []string{"x\ry"}); err == nil {
		t.Error("carriage return in restore workspace accepted")
	}
	if _, err := buildSpawnCommand("a", "", "p\nq", "true\nfalse", nil); err != nil {
		t.Errorf("newline in project or command rejected: %v", err)
	}
}

func TestRestoreTargets(t *testing.T) {
	wks := []i3.Workspace{
		{Name: "1", Output: "A", Visible: true, Focused: true},
		{Name: "2", Output: "A"},
		{Name: "3", Output: "B", Visible: true},
		{Name: "4", Output: "B"},
	}
	tests := []struct {
		output string
		want   []string
	}{
		{"A", []string{"1"}},
		{"", []string{"1"}},
		{"B", []string{"3", "1"}},
	}
	for _, tt := range tests {
		if got := restoreTargets(wks, tt.output); !slices.Equal(got, tt.want) {
			t.Errorf("output %q: got %q, want %q", tt.output, got, tt.want)
		}
	}
	if got := restoreTargets(nil, "A"); got != nil {
		t.Errorf("no focused workspace: got %q", got)
	}
}
