//go:build goexperiment.jsonv2

package configulator

import "fmt"

// ParseError reports a value that could not be parsed into its field.
type ParseError struct {
	// Path is the dotted field path (tag names).
	Path string
	// Source names the input: an env var, a flag, or a file path.
	Source string
	// Value is the raw input value. Fields marked secret: never reach a
	// ParseError with their value; the generator substitutes "(redacted)".
	Value string
	// Err is the underlying parse failure.
	Err error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: cannot parse %q from %s: %v", e.Path, e.Value, e.Source, e.Err)
}
func (e *ParseError) Unwrap() error { return e.Err }

// MissingFileError reports an explicitly named config file that could not
// be read. A path the operator typed is a path the operator expects to
// exist: this fires regardless of FileOptions.RequireFound, and no search
// paths are consulted as a fallback.
type MissingFileError struct {
	// Path is the explicit path as given.
	Path string
	// Searched lists search paths that were configured (informational;
	// they were NOT tried).
	Searched []string
	// Err is the underlying IO error.
	Err error
}

func (e *MissingFileError) Error() string {
	return fmt.Sprintf("config file %s: %v", e.Path, e.Err)
}
func (e *MissingFileError) Unwrap() error { return e.Err }

// SearchPathError reports a configured search path that exists but could
// not be used, i.e. a directory, or unreadable. A merely missing search path
// is a soft miss and does not produce this.
type SearchPathError struct {
	Path string
	Err  error
}

func (e *SearchPathError) Error() string {
	return fmt.Sprintf("config search path %s: %v", e.Path, e.Err)
}
func (e *SearchPathError) Unwrap() error { return e.Err }

// NoFileFoundError reports that RequireFound was set and no search path
// matched.
type NoFileFoundError struct {
	Searched []string
}

func (e *NoFileFoundError) Error() string {
	return fmt.Sprintf("no config file found; searched %v", e.Searched)
}

// RequiredError reports a field tagged required:"true" that no layer set.
type RequiredError struct {
	// Path is the dotted field path.
	Path string
}

func (e *RequiredError) Error() string {
	return fmt.Sprintf("%s: required but not set by any layer", e.Path)
}

// BadEnvOptionsError reports environment options the naming scheme cannot
// survive: a prefix that is not uppercase, or a separator containing "-".
// These are runtime values the generator cannot see, so they are validated
// at Load rather than at generate time. In v1 both failed silently or with
// a misleading "field not found" error.
type BadEnvOptionsError struct {
	Field  string // "Prefix" or "Separator"
	Value  string
	Reason string
}

func (e *BadEnvOptionsError) Error() string {
	return fmt.Sprintf("EnvironmentVariableOptions.%s %q: %s", e.Field, e.Value, e.Reason)
}

// DecodeError wraps a failure from the caller-supplied Unmarshal function,
// naming the file it came from.
type DecodeError struct {
	Path string
	Err  error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("decoding %s: %v", e.Path, e.Err)
}
func (e *DecodeError) Unwrap() error { return e.Err }

// OpaqueSpellingError is returned by a sentinel slot's applier when a
// decoder walked the slot's fields instead of calling UnmarshalText.
// i.e. the config file spelled a text-scalar leaf as a table/mapping.
type OpaqueSpellingError struct {
	Path string
	Hint string // e.g. `"10.0.0.0/8"`
}

func (e *OpaqueSpellingError) Error() string {
	return fmt.Sprintf("%s: expected a text scalar (e.g. %s), got a nested table/mapping", e.Path, e.Hint)
}
