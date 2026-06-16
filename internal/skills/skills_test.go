package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFindsSkillsAndDeduplicatesByRootOrder(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	writeSkill(t, rootA, "alpha", "first")
	writeSkill(t, rootB, "alpha", "second")
	writeSkill(t, rootB, "beta", "second beta")

	reg := Discover("", Config{Enabled: true, Roots: []string{rootA, rootB}})

	if len(reg.Skills) != 2 {
		t.Fatalf("len(Skills) = %d, want 2; errors=%v", len(reg.Skills), reg.Errors)
	}
	alpha, ok := reg.Find("alpha")
	if !ok {
		t.Fatal("alpha not found")
	}
	if alpha.Description != "first" {
		t.Fatalf("alpha description = %q, want first", alpha.Description)
	}
}

func TestDiscoverSkipsMalformedSkills(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "bad")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# Missing frontmatter"), 0o600); err != nil {
		t.Fatal(err)
	}

	reg := Discover("", Config{Enabled: true, Roots: []string{root}})
	if len(reg.Skills) != 0 {
		t.Fatalf("len(Skills) = %d, want 0", len(reg.Skills))
	}
	if len(reg.Errors) == 0 {
		t.Fatal("expected parse error")
	}
}

func TestReadSkillFileRejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha", "first")
	reg := Discover("", Config{Enabled: true, Roots: []string{root}})

	if _, _, err := reg.ReadSkillFile("alpha", "../secret.txt"); err == nil {
		t.Fatal("expected traversal error")
	}
}

func TestSyncMirrorCreatesSymlinks(t *testing.T) {
	root := t.TempDir()
	mirror := t.TempDir()
	writeSkill(t, root, "alpha", "first")
	reg := Discover("", Config{Enabled: true, Roots: []string{root}})

	created, skipped, err := reg.SyncMirror(mirror)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 1 || created[0] != "alpha" || len(skipped) != 0 {
		t.Fatalf("created=%v skipped=%v", created, skipped)
	}
	info, err := os.Lstat(filepath.Join(mirror, "alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("mirror entry is not a symlink")
	}
}

func writeSkill(t *testing.T, root, name, desc string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\ndescription: " + desc + "\n---\n\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
