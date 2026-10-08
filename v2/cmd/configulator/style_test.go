package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// generatedFiles returns the committed generator output: the conformance
// shapes, the package examples and the example programs.
func generatedFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	for _, pattern := range []string{
		"internal/conformance/*_configulator.go",
		"internal/example*/*_configulator.go",
		"examples/*/*_configulator.go",
	} {
		m, err := filepath.Glob(filepath.Join(v2Path(t), pattern))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, m...)
	}
	if len(files) < 10 {
		t.Fatalf("found only %d generated files", len(files))
	}
	return files
}

// TestGeneratedStyle checks the committed generator output against the
// gofmt, gofumpt and goimports rules a linter would report.
func TestGeneratedStyle(t *testing.T) {
	t.Parallel()
	for _, name := range generatedFiles(t) {
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, problem := range styleProblems(src) {
			t.Errorf("%s: %s", name, problem)
		}
	}
}

// styleProblems lists the ways src breaks the style rules generated code
// must follow.
func styleProblems(src []byte) []string {
	var out []string
	formatted, err := format.Source(src)
	if err != nil {
		return []string{err.Error()}
	}
	if !bytes.Equal(formatted, src) {
		out = append(out, "not gofmt-clean")
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return append(out, err.Error())
	}
	out = append(out, importProblems(fset, f)...)
	out = append(out, tagProblems(f)...)
	lines := strings.Split(string(src), "\n")
	for i, d := range f.Decls {
		if i == 0 {
			continue
		}
		start := d.Pos()
		if g, ok := d.(*ast.GenDecl); ok && g.Doc != nil {
			start = g.Doc.Pos()
		}
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Doc != nil {
			start = fn.Doc.Pos()
		}
		if line := fset.Position(start).Line; strings.TrimSpace(lines[line-2]) != "" {
			out = append(out, "no blank line before the declaration on line "+strconv.Itoa(line))
		}
	}
	for _, bad := range []string{"; true {", "WriteString(fmt.Sprintf", "strings.Join([]string{"} {
		if bytes.Contains(src, []byte(bad)) {
			out = append(out, "contains "+bad)
		}
	}
	return out
}

// tagProblems reports a struct whose field tags don't line up: each key
// must start in the same column on every field.
func tagProblems(f *ast.File) []string {
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok {
			return true
		}
		cols := map[string]int{}
		for _, fl := range st.Fields.List {
			if fl.Tag == nil {
				continue
			}
			for _, key := range []string{"toml:", "yaml:"} {
				col := strings.Index(fl.Tag.Value, key)
				if prev, seen := cols[key]; seen && prev != col {
					out = append(out, "unaligned struct tag "+fl.Tag.Value)
				}
				cols[key] = col
			}
		}
		return true
	})
	return out
}

// importProblems checks the goimports layout: standard library imports
// first, then one blank line, then the rest, and no alias that only
// repeats the package name.
func importProblems(fset *token.FileSet, f *ast.File) []string {
	var out []string
	lastStd, firstOther := 0, 0
	prevLine := 0
	for _, imp := range f.Imports {
		p, _ := strconv.Unquote(imp.Path.Value)
		line := fset.Position(imp.Pos()).Line
		std := !strings.Contains(strings.SplitN(p, "/", 2)[0], ".")
		switch {
		case std && firstOther != 0:
			out = append(out, "standard import "+p+" after other imports")
		case std:
			lastStd = line
		case firstOther == 0:
			firstOther = line
		}
		if prevLine != 0 && line != prevLine+1 && line != firstOther {
			out = append(out, "blank line inside an import group before "+p)
		}
		prevLine = line
		if imp.Name != nil && imp.Name.Name == assumedName(p) {
			out = append(out, "redundant import alias "+imp.Name.Name)
		}
	}
	if lastStd != 0 && firstOther != 0 && firstOther != lastStd+2 {
		out = append(out, "standard and other imports are not separated by one blank line")
	}
	return out
}

// assumedName is the package name goimports assumes for an import path:
// the last element, skipping a major version suffix like v2.
func assumedName(p string) string {
	base := path.Base(p)
	if len(base) > 1 && base[0] == 'v' && strings.Trim(base[1:], "0123456789") == "" {
		base = path.Base(path.Dir(p))
	}
	return base
}
