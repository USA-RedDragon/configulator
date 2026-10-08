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
		"yaml": {string(emitSample(m)), "\nbad: []\n"},
		"json": {string(j), `"bad": []`},
		"toml": {string(emitSampleTOML(m)), "bad = []\n"},
	} {
		if !strings.Contains(tc.got, tc.want) {
			t.Errorf("%s sample:\n%s\nwant %q", format, tc.got, tc.want)
		}
	}
	if strings.Contains(string(schema), `"default"`) {
		t.Errorf("schema has a default:\n%s", schema)
	}
}
