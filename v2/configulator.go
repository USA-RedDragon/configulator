//go:build goexperiment.jsonv2

package configulator

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/USA-RedDragon/configulator/v2/internal/seam"
)

// Configulator loads a configuration of type C from defaults, an optional
// config file, environment variables, and an optional flag adapter, in that
// precedence order
type Configulator[C any] struct {
	schema *Schema[C]

	fileOpts *FileOptions
	envOpts  *EnvironmentVariableOptions
	arraySep string

	getenv   Getenv
	readFile func(string) ([]byte, error)

	// Installed by a flag adapter's Bind (type-erased so /v2 names no
	// flag package):
	flagApply      func(*C, SetOrigin) error
	flagConfigPath func() (path string, changed bool)
	bindErr        error

	report *Report
	cfg    *C
}

// New creates a Configulator from a generated Schema (ConfigSchema()).
func New[C any](s *Schema[C]) *Configulator[C] {
	c := &Configulator[C]{
		schema:   s,
		arraySep: ",",
		getenv:   os.LookupEnv,
		readFile: defaultReadFile,
		report:   newReport(),
	}
	registerSeam(c)
	return c
}

// WithFile enables the file layer.
func (c *Configulator[C]) WithFile(opts *FileOptions) *Configulator[C] {
	if opts == nil {
		opts = &FileOptions{}
	}
	c.fileOpts = opts
	return c
}

// WithEnvironmentVariables enables the env layer. A nil opts means no
// prefix and separator "__".
func (c *Configulator[C]) WithEnvironmentVariables(opts *EnvironmentVariableOptions) *Configulator[C] {
	if opts == nil {
		opts = &EnvironmentVariableOptions{}
	}
	if opts.Separator == "" {
		opts.Separator = "__"
	}
	c.envOpts = opts
	return c
}

// WithArraySeparator sets the list separator for the env and defaults
// layers. The CLI layer keeps pflag's comma semantics
func (c *Configulator[C]) WithArraySeparator(sep string) *Configulator[C] {
	c.arraySep = sep
	return c
}

// WithEnviron injects an environment lookup for tests.
func (c *Configulator[C]) WithEnviron(g Getenv) *Configulator[C] {
	c.getenv = g
	return c
}

// WithReadFile injects a file reader for tests.
func (c *Configulator[C]) WithReadFile(f func(string) ([]byte, error)) *Configulator[C] {
	c.readFile = f
	return c
}

// registerSeam hands the adapter view to internal/seam, keyed by this
// Configulator. Flag adapters pop it in Bind. This is the whole flag
// machinery channel: no exported method, no exported type, nothing in
// this package's documentation — v1's you-cannot-misuse-it bar.
func registerSeam[C any](c *Configulator[C]) {
	seam.Register(c, seam.Flag[C, *FileOptions, SetOrigin]{
		FileOptions: func() *FileOptions { return c.fileOpts },
		Install: func(apply func(*C, SetOrigin) error, configPath func() (string, bool), regErr error) {
			c.flagApply = apply
			c.flagConfigPath = configPath
			if regErr != nil && c.bindErr == nil {
				c.bindErr = regErr
			}
		},
	})
}

// Report returns the origin report. Before Load it is empty, never nil.
func (c *Configulator[C]) Report() *Report { return c.report }

// Default returns the configuration with only the defaults layer applied.
func (c *Configulator[C]) Default() (C, error) {
	var cfg C
	if c.schema == nil || c.schema.ApplyDefaults == nil {
		return cfg, fmt.Errorf("configulator: nil Schema or Schema.ApplyDefaults; use the generated ConfigSchema()")
	}
	err := c.schema.ApplyDefaults(&cfg, func(string, Layer, string) {})
	return cfg, err
}

// Load runs every configured layer in precedence order and then Validate,
// if C implements Validator. On error it returns the partially loaded
// config: defaults plus every layer that completed, a layer that fails
// part-way contributes nothing.
func (c *Configulator[C]) Load() (*C, error) {
	cfg, err := c.load()
	if err != nil {
		return cfg, err
	}
	if v, ok := any(cfg).(Validator); ok {
		if err := v.Validate(); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

// LoadWithoutValidation is Load without the Validator step.
func (c *Configulator[C]) LoadWithoutValidation() (*C, error) {
	return c.load()
}

func (c *Configulator[C]) load() (*C, error) {
	if c.schema == nil {
		return nil, fmt.Errorf("configulator: nil Schema; use the generated ConfigSchema()")
	}
	for name, fn := range map[string]bool{
		"ApplyDefaults": c.schema.ApplyDefaults == nil,
		"DecodeFile":    c.schema.DecodeFile == nil,
		"ApplyEnv":      c.schema.ApplyEnv == nil,
	} {
		if fn {
			return nil, fmt.Errorf("configulator: Schema.%s is nil; regenerate with the current configulator version", name)
		}
	}
	if c.bindErr != nil {
		return nil, c.bindErr
	}

	cfg := new(C)
	c.cfg = cfg

	// 1. Defaults.
	set := c.report.set
	if err := c.schema.ApplyDefaults(cfg, set); err != nil {
		return cfg, err
	}

	// 2. File. The flag adapter supplies --config out of band.
	if c.fileOpts != nil || c.flagConfigChanged() {
		opts := c.fileOpts
		if opts == nil {
			opts = &FileOptions{}
		}
		flagPath, flagChanged := "", false
		if c.flagConfigPath != nil {
			flagPath, flagChanged = c.flagConfigPath()
		}
		data, res, err := resolveFile(opts, c.readFile, flagPath, flagChanged)
		c.report.setFile(res)
		if err != nil {
			return cfg, err
		}
		if res.Path != "" && data != nil {
			dec, err := decoderFor(opts.Decoders, res.Path)
			if err != nil {
				return cfg, err
			}
			// The shadow inside DecodeFile is the file layer's stage:
			// applyTo runs only after the whole document decoded, and
			// commits nothing on failure (rule 11).
			staged := *cfg
			if err := c.schema.DecodeFile(data, dec, &staged, set, res.Path); err != nil {
				return cfg, err
			}
			*cfg = staged
		}
	}

	// 3. Environment.
	if c.envOpts != nil {
		if err := validateEnvOptions(c.envOpts); err != nil {
			return cfg, err
		}
		staged := *cfg
		ec := EnvContext{Getenv: c.getenv, Opts: *c.envOpts, ArraySeparator: c.arraySep}
		if err := c.schema.ApplyEnv(&staged, ec, set); err != nil {
			return cfg, err
		}
		*cfg = staged
	}

	// 4. Flags (highest precedence).
	if c.flagApply != nil {
		staged := *cfg
		if err := c.flagApply(&staged, set); err != nil {
			return cfg, err
		}
		*cfg = staged
	}

	// 5. required:"true" checks, after every layer had its chance: a
	// required field must have a recorded origin (a default: tag counts —
	// "required" means "some layer or default supplied it").
	for _, path := range c.schema.Required {
		if _, ok := c.report.Origin(path); !ok {
			return cfg, &RequiredError{Path: path}
		}
	}

	return cfg, nil
}

func (c *Configulator[C]) flagConfigChanged() bool {
	if c.flagConfigPath == nil {
		return false
	}
	_, changed := c.flagConfigPath()
	return changed
}

// validateEnvOptions rejects option values the naming scheme cannot survive
func validateEnvOptions(o *EnvironmentVariableOptions) error {
	if o.Prefix != strings.ToUpper(o.Prefix) {
		return &BadEnvOptionsError{Field: "Prefix", Value: o.Prefix,
			Reason: "must be uppercase; env names are constructed uppercase and a lowercase prefix can never match"}
	}
	if strings.Contains(o.Separator, "-") {
		return &BadEnvOptionsError{Field: "Separator", Value: o.Separator,
			Reason: `must not contain "-"; tag-segment folding would desync the constructed name from the lookup name`}
	}
	return nil
}

type ctxKey struct{}

// WithContext stores the Configulator in a context, mirroring v1.
func (c *Configulator[C]) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// FromContext retrieves a Configulator stored by WithContext.
func FromContext[C any](ctx context.Context) (*Configulator[C], error) {
	c, ok := ctx.Value(ctxKey{}).(*Configulator[C])
	if !ok {
		return nil, fmt.Errorf("configulator: no Configulator[%T] in context", *new(C))
	}
	return c, nil
}
