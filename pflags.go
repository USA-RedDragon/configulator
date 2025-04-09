package configulator

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/USA-RedDragon/configulator/internal/flag"
	inref "github.com/USA-RedDragon/configulator/internal/reflect"
	"github.com/USA-RedDragon/configulator/internal/wrapper"
	"github.com/spf13/pflag"
)

const (
	ConfigFileKey = "config"
)

func (c *Configulator[C]) registerFlags() error {
	if c.flags == nil {
		return fmt.Errorf("flags not set")
	}

	if c.fileOptions != nil && len(c.fileOptions.Paths) > 0 {
		c.flags.StringP(ConfigFileKey, "c", c.fileOptions.Paths[0], "config file")
	}

	var separator string
	if c.pflagOptions != nil {
		separator = c.pflagOptions.Separator
	}
	err := flag.RegisterFlagsFromStruct(c.flags, c.cfg, "", separator, c.arraySeparator)
	if err != nil {
		return fmt.Errorf("failed to register flags: %w", err)
	}
	return nil
}

func (c *Configulator[C]) loadFromFlags() error {
	errs := make([]error, 0)
	wg := sync.WaitGroup{}

	wg.Add(1)
	// Load the flags
	c.flags.VisitAll(func(f *pflag.Flag) {
		wg.Add(1)
		defer wg.Done()
		if f.Name == ConfigFileKey {
			return
		}
		if f.Changed {
			var sep string
			if c.pflagOptions != nil {
				sep = c.pflagOptions.Separator
			}
			flagName := strings.Split(f.Name, sep)
			val := reflect.ValueOf(c.cfg).Elem()
			if len(flagName) < 2 {
				field, err := inref.GetStructFieldByName(reflect.TypeOf(*c.cfg), f.Name, c.arraySeparator)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to get field %s: %w", f.Name, err))
					return
				}
				wrapped, err := wrapper.WrapString(field.Type, f.Value.String(), c.arraySeparator)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to wrap value for field %s: %w", f.Name, err))
					return
				}
				err = inref.SetStructValue(&val, field, wrapped)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to set value for field %s: %w", f.Name, err))
					return
				}
			} else {
				typ, err := inref.GetNestedStructFieldTypeByName(reflect.TypeOf(*c.cfg), flagName, c.arraySeparator)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to get field %s: %w", f.Name, err))
					return
				}
				wrapped, err := wrapper.WrapString(*typ, f.Value.String(), c.arraySeparator)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to wrap value for field %s: %w", f.Name, err))
					return
				}
				err = inref.SetNestedStructValue(&val, flagName, wrapped, c.arraySeparator)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to set value for field %s: %w", f.Name, err))
					return
				}
			}
		}
	})
	wg.Done()
	wg.Wait()
	if len(errs) > 0 {
		return fmt.Errorf("failed to load flags: %v", errs)
	}
	return nil
}
