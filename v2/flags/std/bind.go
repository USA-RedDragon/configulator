//go:build goexperiment.jsonv2

// Package std binds a generated configulator schema to a stdlib
// flag.FlagSet. Stdlib flag has no shorthand, no slice types, and no
// repeated-flag accumulation; the generator's -flags=std output accounts
// for all three
package std

import (
	"flag"
	"fmt"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/configulator/v2/internal/seam"
)

// Options configures flag naming.
type Options struct {
	// Separator joins nested levels in flag names ("." by default,
	// giving --http.port).
	Separator string
}

// Hooks is provided by generated code (-flags=std).
type Hooks[C any] struct {
	// Register adds every config flag to fs.
	Register func(fs *flag.FlagSet, o *Options) error
	// Apply writes every operator-set flag onto the config. isSet holds
	// the names the operator actually set (flag.Visit), stdlib flag's
	// presence oracle.
	Apply func(cfg *C, fs *flag.FlagSet, o *Options, isSet map[string]bool, set configulator.SetOrigin) error
}

// Bind registers the config flags eagerly and installs the apply hook.
// The set map passed to Apply holds the names the operator actually set
// (flag.Visit), stdlib flag's presence oracle.
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
	sm := v.(seam.Flag[C, *configulator.FileOptions, configulator.SetOrigin])
	if fo := sm.FileOptions(); fo != nil {
		if fo.Shorthand != "" {
			regErr = fmt.Errorf("FileOptions.Shorthand %q: stdlib flag has no shorthand concept; unset it or use the pflag adapter", fo.Shorthand)
		} else {
			name := fo.FlagName
			if name == "" {
				name = "config"
			}
			if fs.Lookup(name) != nil {
				regErr = fmt.Errorf("flag -%s already registered on this FlagSet; set FileOptions.FlagName to avoid the collision", name)
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

	apply := func(cfg *C, set configulator.SetOrigin) error {
		if h.Apply == nil {
			return nil
		}
		isSet := map[string]bool{}
		fs.Visit(func(f *flag.Flag) { isSet[f.Name] = true })
		return h.Apply(cfg, fs, o, isSet, set)
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
