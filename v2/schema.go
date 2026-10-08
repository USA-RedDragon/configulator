//go:build goexperiment.jsonv2

package configulator

import "github.com/USA-RedDragon/configulator/v2/impl"

// SetOrigin records where a field's value came from.
type SetOrigin func(path string, layer Layer, detail string)

// Schema is filled in by generated code (ConfigSchema()).
type Schema[C any] struct {
	// ApplyDefaults sets every default: tag. sep is the list separator set
	// by WithArraySeparator, which splits list defaults.
	ApplyDefaults func(cfg *C, sep string, set SetOrigin) error
	// DecodeFile decodes data into a generated shadow struct, then copies
	// the fields present in the file onto cfg. sep splits the list
	// defaults of the elements it builds.
	DecodeFile func(data []byte, u Unmarshal, cfg *C, sep string, set SetOrigin, file string) error
	ApplyEnv   func(cfg *C, ec EnvContext, set SetOrigin) error
	// Required lists the dotted paths of required:"true" fields outside
	// optional structs.
	Required []string
	// ConditionalRequired returns the dotted paths of required:"true"
	// fields whose container exists in cfg: an optional struct some layer
	// allocated, or an element of a list or map. Nil if there are none.
	ConditionalRequired func(cfg *C) []string
}

// EnvContext is what the generated env code needs at Load.
type EnvContext struct {
	Getenv         Getenv
	Opts           EnvironmentVariableOptions
	ArraySeparator string
}

// Lookup builds the env var name for a field from its name segments, the
// same way EnvName does, and looks it up. It returns the name even when the
// variable is not set.
func (ec EnvContext) Lookup(segments ...string) (name, value string, ok bool) {
	name = impl.EnvName(ec.Opts.Prefix, ec.Opts.Separator, segments...)
	value, ok = ec.Getenv(name)
	return name, value, ok
}

// Validator is called by Load when the config type implements it.
type Validator interface {
	Validate() error
}
