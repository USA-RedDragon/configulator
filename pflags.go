package configulator

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	inref "github.com/USA-RedDragon/configulator/internal/reflect"
	"github.com/USA-RedDragon/configulator/internal/tags"
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

	err := c.registerFlagsFromStruct(c.cfg, "")
	if err != nil {
		return fmt.Errorf("failed to register flags: %w", err)
	}
	return nil
}

func (c *Configulator[C]) registerFlagsFromStruct(stru any, prefix string) error {
	typ := reflect.TypeOf(stru).Elem()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", typ.Kind())
	}

	for i := range typ.NumField() {
		field := typ.Field(i)
		if tag := field.Tag.Get("config"); tag != "" {
			err := c.addFlag(prefix+tag, field)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Configulator[C]) addFlag(tagStr string, field reflect.StructField) error {
	tag, err := tags.ExtractStructTag(field, tagStr, c.arraySeparator)
	if err != nil {
		return err
	}

	switch field.Type.Kind() {
	case reflect.Bool:
		c.flags.Bool(tag.Name, tag.DefaultVal.UnwrapBool(), tag.Description)
	case reflect.Int8:
		c.flags.Int8(tag.Name, tag.DefaultVal.UnwrapInt8(), tag.Description)
	case reflect.Int16:
		c.flags.Int16(tag.Name, tag.DefaultVal.UnwrapInt16(), tag.Description)
	case reflect.Int32:
		c.flags.Int32(tag.Name, tag.DefaultVal.UnwrapInt32(), tag.Description)
	case reflect.Int64:
		c.flags.Int64(tag.Name, tag.DefaultVal.UnwrapInt64(), tag.Description)
	case reflect.Int:
		c.flags.Int(tag.Name, tag.DefaultVal.UnwrapInt(), tag.Description)
	case reflect.Uint8:
		c.flags.Uint8(tag.Name, tag.DefaultVal.UnwrapUint8(), tag.Description)
	case reflect.Uint16:
		c.flags.Uint16(tag.Name, tag.DefaultVal.UnwrapUint16(), tag.Description)
	case reflect.Uint32:
		c.flags.Uint32(tag.Name, tag.DefaultVal.UnwrapUint32(), tag.Description)
	case reflect.Uint64:
		c.flags.Uint64(tag.Name, tag.DefaultVal.UnwrapUint64(), tag.Description)
	case reflect.Uint:
		c.flags.Uint(tag.Name, tag.DefaultVal.UnwrapUint(), tag.Description)
	case reflect.Float32:
		c.flags.Float32(tag.Name, tag.DefaultVal.UnwrapFloat32(), tag.Description)
	case reflect.Float64:
		c.flags.Float64(tag.Name, tag.DefaultVal.UnwrapFloat64(), tag.Description)
	case reflect.String:
		c.flags.String(tag.Name, tag.DefaultVal.UnwrapString(), tag.Description)
	case reflect.Struct:
		sep := "."
		if c.pflagOptions != nil && c.pflagOptions.Separator != "" {
			sep = c.pflagOptions.Separator
		}
		c.registerFlagsFromStruct(reflect.New(field.Type).Interface(), tag.Name+sep)
	case reflect.Array, reflect.Slice:
		switch field.Type.Elem().Kind() {
		case reflect.Bool:
			c.flags.BoolSlice(tag.Name, tag.DefaultVal.UnwrapBoolSlice(), tag.Description)
		case reflect.Float32:
			c.flags.Float32Slice(tag.Name, tag.DefaultVal.UnwrapFloat32Slice(), tag.Description)
		case reflect.Float64:
			c.flags.Float64Slice(tag.Name, tag.DefaultVal.UnwrapFloat64Slice(), tag.Description)
		case reflect.Int32:
			c.flags.Int32Slice(tag.Name, tag.DefaultVal.UnwrapInt32Slice(), tag.Description)
		case reflect.Int64:
			c.flags.Int64Slice(tag.Name, tag.DefaultVal.UnwrapInt64Slice(), tag.Description)
		case reflect.Int:
			c.flags.IntSlice(tag.Name, tag.DefaultVal.UnwrapIntSlice(), tag.Description)
		case reflect.String:
			c.flags.StringSlice(tag.Name, tag.DefaultVal.UnwrapStringSlice(), tag.Description)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			c.flags.UintSlice(tag.Name, tag.DefaultVal.UnwrapUintSlice(), tag.Description)
		default:
			return fmt.Errorf("unsupported type in config: %v", field.Type)
		}
	default:
		return fmt.Errorf("unsupported type in config: %v", field.Type)
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
			sep := "."
			if c.pflagOptions != nil && c.pflagOptions.Separator != "" {
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
