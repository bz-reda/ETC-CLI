package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFramework(t *testing.T) {
	cases := []struct{ file, content, want string }{
		{"go.mod", "module x", "go"},
		{"requirements.txt", "flask", "python"},
		{"composer.json", "{}", "php"},
		{"Gemfile", "source 'x'", "ruby"},
		{"Cargo.toml", "[package]", "rust"},
		{"index.html", "<h1>x</h1>", "static"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, c.file), []byte(c.content), 0644)
		if got := detectFramework(dir); got != c.want {
			t.Errorf("%s ⇒ %q, want %q", c.file, got, c.want)
		}
	}
	// package.json with next ⇒ nextjs; without ⇒ node
	d1 := t.TempDir()
	os.WriteFile(filepath.Join(d1, "package.json"), []byte(`{"dependencies":{"next":"15"}}`), 0644)
	if detectFramework(d1) != "nextjs" {
		t.Error("package.json with next must detect nextjs")
	}
	d2 := t.TempDir()
	os.WriteFile(filepath.Join(d2, "package.json"), []byte(`{"dependencies":{}}`), 0644)
	if detectFramework(d2) != "node" {
		t.Error("package.json without next must detect node")
	}
	// nothing recognized ⇒ auto
	if detectFramework(t.TempDir()) != "auto" {
		t.Error("empty dir must detect auto")
	}
}
