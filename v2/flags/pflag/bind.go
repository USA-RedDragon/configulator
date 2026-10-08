//go:build goexperiment.jsonv2

// Package pflag binds a generated configulator schema to a
// github.com/spf13/pflag FlagSet.
package pflag

import (
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/configulator/v2/internal/seam"
	"github.com/spf13/pflag"
)

// Options configures flag naming.
type Options struct {
	// Separator joins nested levels in flag names. Default "." (--http.port).
	Separator string
}

// Hooks is provided by generated code (-flags=pflag).
type Hooks[C any] struct {
	// Register adds every config flag to fs. It returns a
	// *configulator.FlagConflictError on a duplicate name or shorthand,
	// since pflag would panic.
	Register func(fs *pflag.FlagSet, o *Options) error
	// Apply copies every changed flag onto the config. sep is the list
	// separator for the defaults of an optional struct a flag allocates.
	Apply func(cfg *C, fs *pflag.FlagSet, o *Options, sep string, set configulator.SetOrigin) error
}

// Bind adds the config flags to fs right away (cobra parses args before
// RunE) and applies them at Load. Call it after WithFile so --config gets
// the right default.
func Bind[C any](c *configulator.Configulator[C], fs *pflag.FlagSet, h Hooks[C], o *Options) *configulator.Configulator[C] {
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
		name := fo.FlagName
		if name == "" {
			name = "config"
		}
		short := fo.Shorthand
		if short == "" {
			short = "c"
		}
		def := ""
		if len(fo.Search) > 0 {
			def = fo.Search[0]
		}
		switch {
		case len(short) != 1:
			regErr = fmt.Errorf("FileOptions.Shorthand %q must be a single ASCII character", short)
		case fs.Lookup(name) != nil:
			regErr = &configulator.FlagConflictError{Flag: name, Existing: fs.Lookup(name).Name}
		case fs.ShorthandLookup(short) != nil:
			regErr = &configulator.FlagConflictError{Flag: name, Shorthand: short, Existing: fs.ShorthandLookup(short).Name}
		default:
			fs.StringP(name, short, def, "config file")
			configFlag = name
		}
	}

	if regErr == nil && h.Register != nil {
		regErr = h.Register(fs, o)
	}

	apply := func(cfg *C, sep string, set configulator.SetOrigin) error {
		if h.Apply == nil {
			return nil
		}
		return h.Apply(cfg, fs, o, sep, set)
	}
	configPath := func() (string, bool) {
		if configFlag == "" {
			return "", false
		}
		f := fs.Lookup(configFlag)
		if f == nil {
			return "", false
		}
		return f.Value.String(), f.Changed
	}
	sm.Install(apply, configPath, regErr)
	return c
}
