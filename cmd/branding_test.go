package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cmdSourceFiles returns the non-test .go files in the cmd package. The
// branding pins scan source, not tests — test files intentionally reference
// the legacy .espacetech.json / .espacetechignore filenames for back-compat.
func cmdSourceFiles(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read cmd dir: %v", err)
	}
	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, name)
	}
	if len(files) == 0 {
		t.Fatal("no cmd source files found")
	}
	return files
}

// TestNoLegacyCommandHints pins that no user-facing command hint references the
// old "espacetech " command name. The trailing SPACE is load-bearing: it
// matches hints like "espacetech login" but never the kept config filenames
// ".espacetech.json" / ".espacetechignore" (no space after the name).
func TestNoLegacyCommandHints(t *testing.T) {
	for _, name := range cmdSourceFiles(t) {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := string(data)
		for i, line := range strings.Split(src, "\n") {
			if strings.Contains(line, "espacetech ") {
				t.Errorf("%s:%d still has a legacy 'espacetech ' command hint: %s",
					name, i+1, strings.TrimSpace(line))
			}
		}
	}
}

// TestNoLegacyBrandName pins that the old "Espace-Tech" brand string is gone
// from cmd source (Short/Long descriptions etc.).
func TestNoLegacyBrandName(t *testing.T) {
	for _, name := range cmdSourceFiles(t) {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(data), "Espace-Tech") {
			t.Errorf("%s still contains legacy brand literal %q", name, "Espace-Tech")
		}
	}
}

// TestLoginHintRebranded is a positive pin: the login hint must use the new
// "ghayma login" command name. Guards against an over-eager edit that drops
// the hint entirely instead of rebranding it.
func TestLoginHintRebranded(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(".", "init.go"))
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	if !strings.Contains(string(data), "ghayma login") {
		t.Errorf("init.go login hint not rebranded to 'ghayma login'")
	}
}
