//go:build goexperiment.jsonv2

package pflag_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	configulator "github.com/USA-RedDragon/configulator/v2"
	cpflag "github.com/USA-RedDragon/configulator/v2/flags/pflag"
	"github.com/USA-RedDragon/configulator/v2/internal/exampleconfig"
	"github.com/spf13/pflag"
)

// TestConflictNamesExistingFlag checks that FlagConflictError.Existing is
// the name of the flag already on the FlagSet, which a NormalizeFunc can
// make different from the name being added.
func TestConflictNamesExistingFlag(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ existing, flag string }{
		{"my.config", "my_config"},
		{"http.port", "http_port"},
	} {
		fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
		fs.SetNormalizeFunc(func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
			return pflag.NormalizedName(strings.ReplaceAll(name, "_", "."))
		})
		fs.String(tc.existing, "", "")
		c := configulator.New(exampleconfig.ConfigSchema()).WithFile(&configulator.FileOptions{FlagName: "my_config"})
		cpflag.Bind(c, fs, exampleconfig.ConfigPFlagHooks(), &cpflag.Options{Separator: "_"})
		_, err := c.Load()
		var fc *configulator.FlagConflictError
		if !errors.As(err, &fc) {
			t.Fatalf("%s: got %v, want a FlagConflictError", tc.existing, err)
		}
		if fc.Flag != tc.flag || fc.Existing != tc.existing {
			t.Errorf("got Flag %q Existing %q, want %q and %q", fc.Flag, fc.Existing, tc.flag, tc.existing)
		}
	}
}

// TestConfigFlagCompletion checks that --config carries cobra's filename
// extension annotation, built from the registered decoders.
func TestConfigFlagCompletion(t *testing.T) {
	t.Parallel()
	const key = "cobra_annotation_bash_completion_filename_extensions"
	yaml := func([]byte, any) error { return nil }
	for _, tc := range []struct {
		decoders configulator.Decoders
		want     []string
	}{
		{nil, []string{"json"}},
		{configulator.Decoders{".yml": yaml, ".yaml": yaml, ".json": configulator.StrictJSON}, []string{"json", "yaml", "yml"}},
		{configulator.Decoders{}, nil},
	} {
		fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
		c := configulator.New(exampleconfig.ConfigSchema()).WithFile(&configulator.FileOptions{Decoders: tc.decoders})
		cpflag.Bind(c, fs, exampleconfig.ConfigPFlagHooks(), nil)
		if got := fs.Lookup("config").Annotations[key]; !slices.Equal(got, tc.want) {
			t.Errorf("decoders %v: annotation %q, want %q", tc.decoders, got, tc.want)
		}
	}
}
