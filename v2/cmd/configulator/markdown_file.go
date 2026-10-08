package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const (
	markdownBegin = "<!-- configulator:begin -->"
	markdownEnd   = "<!-- configulator:end -->"
)

var errStale = errors.New("is out of date")

// spliceMarkdown replaces everything between the begin and end markers in doc
// with table. The markers themselves are kept.
func spliceMarkdown(doc, table []byte) ([]byte, error) {
	begin := bytes.Index(doc, []byte(markdownBegin))
	end := bytes.Index(doc, []byte(markdownEnd))
	if begin < 0 || end < 0 || end < begin {
		return nil, fmt.Errorf("needs a %s line followed by a %s line", markdownBegin, markdownEnd)
	}
	if bytes.Contains(doc[end+len(markdownEnd):], []byte(markdownBegin)) {
		return nil, fmt.Errorf("has more than one %s marker", markdownBegin)
	}
	var out bytes.Buffer
	out.Write(doc[:begin+len(markdownBegin)])
	out.WriteString("\n\n")
	out.Write(bytes.TrimRight(table, "\n"))
	out.WriteString("\n\n")
	out.Write(doc[end:])
	return out.Bytes(), nil
}

// updateMarkdownFile writes table into path between the markers. With check
// set it writes nothing and returns errStale if the file would change.
func updateMarkdownFile(path string, table []byte, check bool) (changed bool, err error) {
	doc, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	out, err := spliceMarkdown(doc, table)
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}
	if bytes.Equal(doc, out) {
		return false, nil
	}
	if check {
		return true, fmt.Errorf("%s %w; run configulator -markdown -markdown-file %s", path, errStale, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, out, info.Mode().Perm())
}

// updateSampleFile replaces path with content, creating it if needed. With
// check set it writes nothing and returns errStale if the file would change.
func updateSampleFile(path string, content []byte, check bool) (changed bool, err error) {
	old, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return false, err
	case bytes.Equal(old, content):
		return false, nil
	}
	if check {
		return true, fmt.Errorf("%s %w; run configulator -sample -sample-file %s", path, errStale, path)
	}
	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(path); statErr == nil {
		mode = info.Mode().Perm()
	}
	return true, os.WriteFile(path, content, mode)
}
