package safepath

import (
	"path/filepath"
	"testing"
)

// A JSON string escape right after a repo-level worktrees path is the escape,
// not a separator plus a run id, so the next word survives.
func TestRedactWorktrees_KeepsJSONEscapeAfterRepoLevelPath(t *testing.T) {
	worktrees := filepath.Join(t.TempDir(), "worktrees")
	for _, esc := range []string{`\n`, `\r`, `\t`} {
		text := filepath.Join(worktrees, "ab12") + esc + "next"
		if got, want := RedactWorktrees(text, "", worktrees), WorktreesPlaceholder+esc+"next"; got != want {
			t.Errorf("RedactWorktrees(%q) = %q, want %q", text, got, want)
		}
	}
}
