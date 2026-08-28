//go:build goexperiment.jsonv2

package configulator

// SetOrigin records one field's origin. Generated appliers receive it as
// a closure
type SetOrigin func(path string, layer Layer, detail string)

// Schema is filled in entirely by generated code (ConfigSchema()).
type Schema[C any] struct {
	// ApplyDefaults assigns every default: tag as a literal.
	ApplyDefaults func(*C, SetOrigin) error
	// DecodeFile decodes data into the hidden shadow via u, then applies
	// present fields onto the config. The shadow type never escapes.
	DecodeFile func(data []byte, u Unmarshal, cfg *C, set SetOrigin, file string) error
	// ApplyEnv reads variables through ec.Getenv and applies present ones.
	ApplyEnv func(cfg *C, ec EnvContext, set SetOrigin) error
	// Required lists the dotted paths of required:"true" fields. Load
	// verifies each has a recorded origin after all layers ran — pure
	// data, no generated logic to misuse.
	Required []string
}

// EnvContext carries everything the generated env applier needs
type EnvContext struct {
	Getenv         Getenv
	Opts           EnvironmentVariableOptions
	ArraySeparator string
}

// Validator is optional at runtime
type Validator interface {
	Validate() error
}
