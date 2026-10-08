//go:build goexperiment.jsonv2

package configulator

// SetOrigin records where a field's value came from.
type SetOrigin func(path string, layer Layer, detail string)

// Schema is filled in by generated code (ConfigSchema()).
type Schema[C any] struct {
	ApplyDefaults func(*C, SetOrigin) error
	// DecodeFile decodes data into a generated shadow struct, then copies
	// the fields present in the file onto cfg.
	DecodeFile func(data []byte, u Unmarshal, cfg *C, set SetOrigin, file string) error
	ApplyEnv   func(cfg *C, ec EnvContext, set SetOrigin) error
	// Required lists the dotted paths of required:"true" fields.
	Required []string
}

// EnvContext is what the generated env code needs at Load.
type EnvContext struct {
	Getenv         Getenv
	Opts           EnvironmentVariableOptions
	ArraySeparator string
}

// Validator is called by Load when the config type implements it.
type Validator interface {
	Validate() error
}
