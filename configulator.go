package configulator

import (
	"errors"
	"fmt"
	"reflect"

	inref "github.com/USA-RedDragon/configulator/internal/reflect"
	"github.com/spf13/pflag"
)

// Config is the interface that must be implemented by the configuration struct.
type Config interface {
	// Validate validates the configuration. It will be called after loading the configuration.
	// If the configuration is invalid, it should return an error.
	Validate() error
}

// Configulator is a configuration loader.
// It loads configuration from environment variables, files, and flags.
// The configuration is loaded in the following order:
// 1. Files
// 2. Environment variables
// 3. Flags
type Configulator[C Config] struct {
	cfg            *C
	envOptions     *EnvironmentVariableOptions
	fileOptions    *FileOptions
	flags          *pflag.FlagSet
	pflagOptions   *PFlagOptions
	arraySeparator string
}

// NewConfigulator creates a new Configulator.
// The defaults are the default configuration values.
func NewConfigulator[C Config]() *Configulator[C] {
	c := &Configulator[C]{
		arraySeparator: ",",
	}
	return c
}

// Load reads the configuration from the environment variables, files, and flags.
func (c *Configulator[C]) Load() (*C, error) {
	cfg, err := c.Default()
	if err != nil {
		return c.cfg, fmt.Errorf("failed to get defaults: %w", err)
	}
	c.cfg = &cfg
	if c.fileOptions != nil || (c.flags != nil && c.flags.Changed(ConfigFileKey)) {
		if c.flags.Changed(ConfigFileKey) {
			c.fileOptions.Paths = []string{c.flags.Lookup(ConfigFileKey).Value.String()}
		}
		err = c.loadFromFile()
		if err != nil {
			if errors.Is(err, ErrConfigFileNotFound) {
				if c.fileOptions.ErrorIfNotFound {
					return c.cfg, fmt.Errorf("failed to load from file: %w", err)
				}
				// file not found but not required
				// continue loading from environment and flags
			} else {
				return c.cfg, fmt.Errorf("failed to load from file: %w", err)
			}
		}
	}
	if c.envOptions != nil {
		err = c.loadFromEnvironment()
		if err != nil {
			return c.cfg, fmt.Errorf("failed to load from environment: %w", err)
		}
	}
	if c.flags != nil {
		err = c.loadFromFlags()
		if err != nil {
			return c.cfg, fmt.Errorf("failed to load from flags: %w", err)
		}
	}

	return c.cfg, (*c.cfg).Validate()
}

// Default returns the default configuration.
func (c *Configulator[C]) Default() (C, error) {
	defs, err := inref.GetDefaultsFromStruct(reflect.TypeFor[C](), c.arraySeparator)
	if err != nil {
		var zero C
		return zero, fmt.Errorf("failed to get defaults: %w", err)
	}
	config, ok := defs.(C)
	if !ok {
		var zero C
		return zero, fmt.Errorf("failed to get defaults")
	}
	return config, nil
}
