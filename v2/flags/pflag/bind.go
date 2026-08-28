//go:build goexperiment.jsonv2

// Package pflag binds a generated configulator schema to a
// github.com/spf13/pflag FlagSet. It is the only place the pflag
// dependency appears: configulator/v2's core names no flag type.
package pflag

import (
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/configulator/v2/internal/seam"
	"github.com/spf13/pflag"
)

// Options configures flag naming.
type Options struct {
	// Separator joins nested levels in flag names ("." by default,
	// giving --http.port).
	Separator string
}

// Hooks is provided by generated code (-flags=pflag).
type Hooks[C any] struct {
	// Register adds every config flag to fs. Generated code checks
	// fs.Lookup before each AddFlag (pflag panics on duplicates) and
	// returns a descriptive error instead.
	Register func(fs *pflag.FlagSet, o *Options) error
	// Apply writes every Changed flag onto the config.
	Apply func(cfg *C, fs *pflag.FlagSet, o *Options, set configulator.SetOrigin) error
}

// Bind registers the config flags eagerly (cobra prints --help and parses
// args before RunE runs, so deferred registration would make every config
// flag an unknown-flag error) and installs the apply hook to run at Load.
//
// Call Bind after WithFile so the --config flag's help text shows the
// right default; the flag itself works regardless of order.
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
	sm := v.(seam.Flag[C, *configulator.FileOptions, configulator.SetOrigin])
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
		if fs.Lookup(name) != nil {
			regErr = fmt.Errorf("flag --%s already registered on this FlagSet; set FileOptions.FlagName to avoid the collision", name)
		} else {
			fs.StringP(name, short, def, "config file")
			configFlag = name
		}
	}

	if regErr == nil && h.Register != nil {
		regErr = h.Register(fs, o)
	}

	apply := func(cfg *C, set configulator.SetOrigin) error {
		if h.Apply == nil {
			return nil
		}
		return h.Apply(cfg, fs, o, set)
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
