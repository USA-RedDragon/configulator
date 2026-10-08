package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpliceMarkdown(t *testing.T) {
	t.Parallel()
	doc := "# App\n\n## Config\n\n" + markdownBegin + "\nold table\n" + markdownEnd + "\n\nFooter\n"
	got, err := spliceMarkdown([]byte(doc), []byte("| new |\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := "# App\n\n## Config\n\n" + markdownBegin + "\n\n| new |\n\n" + markdownEnd + "\n\nFooter\n"
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}

	again, err := spliceMarkdown(got, []byte("| new |\n"))
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(got) {
		t.Error("splicing the same table twice changed the document")
	}

	for name, bad := range map[string]string{
		"no markers":   "# App\n",
		"only begin":   markdownBegin + "\n",
		"reversed":     markdownEnd + "\n" + markdownBegin + "\n",
		"two sections": markdownBegin + markdownEnd + markdownBegin + markdownEnd,
	} {
		if _, err := spliceMarkdown([]byte(bad), []byte("x")); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestUpdateMarkdownFileCheck(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(path, []byte(markdownBegin+"\n"+markdownEnd+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	changed, err := updateMarkdownFile(path, []byte("| t |"), true)
	if !changed || !errors.Is(err, errStale) {
		t.Fatalf("check on a stale file: changed=%v err=%v", changed, err)
	}
	if b, _ := os.ReadFile(path); strings.Contains(string(b), "| t |") {
		t.Fatal("check mode wrote the file")
	}

	if _, err := updateMarkdownFile(path, []byte("| t |"), false); err != nil {
		t.Fatal(err)
	}
	changed, err = updateMarkdownFile(path, []byte("| t |"), true)
	if changed || err != nil {
		t.Fatalf("check on an up-to-date file: changed=%v err=%v", changed, err)
	}
}

func TestUpdateSampleFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.example.yaml")

	changed, err := updateSampleFile(path, []byte("a: 1\n"), true)
	if !changed || !errors.Is(err, errStale) {
		t.Fatalf("check on a missing file: changed=%v err=%v", changed, err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("check mode created the file")
	}

	if changed, err = updateSampleFile(path, []byte("a: 1\n"), false); !changed || err != nil {
		t.Fatalf("create: changed=%v err=%v", changed, err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	if changed, err = updateSampleFile(path, []byte("a: 1\n"), true); changed || err != nil {
		t.Fatalf("check on an up-to-date file: changed=%v err=%v", changed, err)
	}
	if changed, err = updateSampleFile(path, []byte("a: 2\n"), false); !changed || err != nil {
		t.Fatalf("update: changed=%v err=%v", changed, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != "a: 2\n" || info.Mode().Perm() != 0o640 {
		t.Fatalf("got %q mode %v", b, info.Mode().Perm())
	}
}
