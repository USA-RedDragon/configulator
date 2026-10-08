//go:build goexperiment.jsonv2

// Package std binds a generated configulator schema to a stdlib
// flag.FlagSet. The stdlib flag package has no shorthands, slice types or
// repeated flags, and the generated -flags=std code handles that.
package std

import (
	"flag"
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/configulator/v2/internal/seam"
)

// Options configures flag naming.
type Options struct {
	// Separator joins nested levels in flag names. Default "." (-http.port).
	Separator string
}

// Hooks is provided by generated code (-flags=std).
type Hooks[C any] struct {
	// Register adds every config flag to fs. It returns a
	// *configulator.FlagConflictError on a duplicate name, since flag
	// would panic.
	Register func(fs *flag.FlagSet, o *Options) error
	// Apply copies the set flags onto the config. isSet holds the names
	// that were set on the command line (from flag.Visit). sep is the list
	// separator for the defaults of an optional struct a flag allocates.
	Apply func(cfg *C, fs *flag.FlagSet, o *Options, isSet map[string]bool, sep string, set configulator.SetOrigin) error
}

// Bind adds the config flags to fs and applies them at Load.
func Bind[C any](c *configulator.Configulator[C], fs *flag.FlagSet, h Hooks[C], o *Options) *configulator.Configulator[C] {
	if o == nil {
		o = &Options{}
	}
	if o.Separator == "" {
		o.Separator = "."
	}

	var regErr error
	configFlag := ""
	v, ok := seam.Take(c)
	if !ok {
		panic("configulator: Bind called twice on the same Configulator, or on one not created by New")
	}
	sm, ok := v.(seam.Flag[C, *configulator.FileOptions, configulator.SetOrigin])
	if !ok {
		panic("configulator: Bind found flag hooks for a different config type")
	}
	if fo := sm.FileOptions(); fo != nil {
		if fo.Shorthand != "" {
			regErr = fmt.Errorf("FileOptions.Shorthand %q: stdlib flag has no shorthand concept; unset it or use the pflag adapter", fo.Shorthand)
		} else {
			name := fo.FlagName
			if name == "" {
				name = "config"
			}
			if f := fs.Lookup(name); f != nil {
				regErr = &configulator.FlagConflictError{Flag: name, Existing: f.Name}
			} else {
				def := ""
				if len(fo.Search) > 0 {
					def = fo.Search[0]
				}
				fs.String(name, def, "config file")
				configFlag = name
			}
		}
	}

	if regErr == nil && h.Register != nil {
		regErr = h.Register(fs, o)
	}

	apply := func(cfg *C, sep string, set configulator.SetOrigin) error {
		if h.Apply == nil {
			return nil
		}
		isSet := map[string]bool{}
		fs.Visit(func(f *flag.Flag) { isSet[f.Name] = true })
		return h.Apply(cfg, fs, o, isSet, sep, set)
	}
	configPath := func() (string, bool) {
		if configFlag == "" {
			return "", false
		}
		set := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == configFlag {
				set = true
			}
		})
		if !set {
			return "", false
		}
		return fs.Lookup(configFlag).Value.String(), true
	}
	sm.Install(apply, configPath, regErr)
	return c
}
