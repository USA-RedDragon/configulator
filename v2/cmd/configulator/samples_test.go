package main

import (
	"strings"
	"testing"
)

// TestSampleBoolListNormalized checks that YAML samples spell bool list
// elements the way the schema, JSON and TOML samples do.
func TestSampleBoolListNormalized(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\ntype Cfg struct {\n"+
		"\tBools []bool `name:\"bools\" default:\"1,f,True\"`\n"+
		"}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(emitSample(m)), "bools: [true, false, true]\n"; !strings.Contains(got, want) {
		t.Errorf("yaml sample:\n%s\nwant line %q", got, want)
	}
}

// TestSampleUnparseableListDefault checks that a list default that doesn't
// parse when split on "," is written as an empty list in every sample
// format, and left out of the schema. It may still be valid at load with
// another WithArraySeparator, so the generator accepts it.
func TestSampleUnparseableListDefault(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\ntype Cfg struct {\n"+
		"\tBad []int `name:\"bad\" default:\"1;x\"`\n"+
		"}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	j, err := emitSampleJSON(m)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := emitJSONSchema(m)
	if err != nil {
		t.Fatal(err)
	}
	for format, tc := range map[string]struct{ got, want string }{
		formatYAML: {string(emitSample(m)), "\nbad: []\n"},
		formatJSON: {string(j), `"bad": []`},
		formatTOML: {string(emitSampleTOML(m)), "bad = []\n"},
	} {
		if !strings.Contains(tc.got, tc.want) {
			t.Errorf("%s sample:\n%s\nwant %q", format, tc.got, tc.want)
		}
	}
	if strings.Contains(string(schema), `"default"`) {
		t.Errorf("schema has a default:\n%s", schema)
	}
}

// TestTOMLTooLargeCommentQuotesKey checks that the comment on a uint64 too
// large for TOML quotes the key, so a key holding a newline can't break
// the file.
func TestTOMLTooLargeCommentQuotesKey(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\ntype Cfg struct {\n"+
		"\tBig uint64 `name:\"a\\nb\" default:\"18446744073709551615\"`\n"+
		"\tBigs []uint64 `name:\"c\\nd\" default:\"18446744073709551615\"`\n"+
		"}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	got := string(emitSampleTOML(m))
	for _, want := range []string{
		"# \"a\\nb\" is too large for a TOML integer, so it's written as a string\n",
		"# \"c\\nd\" holds a number too large for a TOML integer, written as a string\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("toml sample:\n%s\nwant %q", got, want)
		}
	}
}

// TestMarkdownEscapesControlChars checks that a newline or tab in a cell
// is written as an escape, so it can't break the table row.
func TestMarkdownEscapesControlChars(t *testing.T) {
	t.Parallel()
	m, err := buildFixtureModel(t, "package fixture\n\ntype Cfg struct {\n"+
		"\tS string `name:\"s\" default:\"a\\nb\\tc\" description:\"one\\ntwo\"`\n"+
		"}\n"+validateStub, false)
	if err != nil {
		t.Fatal(err)
	}
	md := string(emitMarkdown(m, ".", "", "_", false))
	if lines := strings.Count(md, "\n"); lines != 3 {
		t.Errorf("want a header, a rule and one row, got %d lines:\n%s", lines, md)
	}
	for _, want := range []string{"`a\\nb\\tc`", "one\\ntwo"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q:\n%s", want, md)
		}
	}
}
