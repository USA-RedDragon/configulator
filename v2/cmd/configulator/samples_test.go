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
