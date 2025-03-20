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
		if tag := field.Tag.Get("name"); tag != "" {
			err := c.addFlag(prefix, field)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Configulator[C]) addFlag(prefix string, field reflect.StructField) error {
	tag, err := tags.ExtractStructTags(field, c.arraySeparator)
	if err != nil {
		return err
	}

	tag.Name = prefix + tag.Name

	switch field.Type.Kind() {
	case reflect.Bool:
		if tag.DefaultVal.Value == nil {
			c.flags.Bool(tag.Name, false, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapBool()
		if !ok {
			return fmt.Errorf("failed to unwrap bool")
		}
		c.flags.Bool(tag.Name, v, tag.Description)
	case reflect.Int8:
		if tag.DefaultVal.Value == nil {
			c.flags.Int8(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt8()
		if !ok {
			return fmt.Errorf("failed to unwrap int8")
		}
		c.flags.Int8(tag.Name, v, tag.Description)
	case reflect.Int16:
		if tag.DefaultVal.Value == nil {
			c.flags.Int16(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt16()
		if !ok {
			return fmt.Errorf("failed to unwrap int16")
		}
		c.flags.Int16(tag.Name, v, tag.Description)
	case reflect.Int32:
		if tag.DefaultVal.Value == nil {
			c.flags.Int32(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt32()
		if !ok {
			return fmt.Errorf("failed to unwrap int32")
		}
		c.flags.Int32(tag.Name, v, tag.Description)
	case reflect.Int64:
		if tag.DefaultVal.Value == nil {
			c.flags.Int64(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt64()
		if !ok {
			return fmt.Errorf("failed to unwrap int64")
		}
		c.flags.Int64(tag.Name, v, tag.Description)
	case reflect.Int:
		if tag.DefaultVal.Value == nil {
			c.flags.Int(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt()
		if !ok {
			return fmt.Errorf("failed to unwrap int")
		}
		c.flags.Int(tag.Name, v, tag.Description)
	case reflect.Uint8:
		if tag.DefaultVal.Value == nil {
			c.flags.Uint8(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint8()
		if !ok {
			return fmt.Errorf("failed to unwrap uint8")
		}
		c.flags.Uint8(tag.Name, v, tag.Description)
	case reflect.Uint16:
		if tag.DefaultVal.Value == nil {
			c.flags.Uint16(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint16()
		if !ok {
			return fmt.Errorf("failed to unwrap uint16")
		}
		c.flags.Uint16(tag.Name, v, tag.Description)
	case reflect.Uint32:
		if tag.DefaultVal.Value == nil {
			c.flags.Uint32(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint32()
		if !ok {
			return fmt.Errorf("failed to unwrap uint32")
		}
		c.flags.Uint32(tag.Name, v, tag.Description)
	case reflect.Uint64:
		if tag.DefaultVal.Value == nil {
			c.flags.Uint64(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint64()
		if !ok {
			return fmt.Errorf("failed to unwrap uint64")
		}
		c.flags.Uint64(tag.Name, v, tag.Description)
	case reflect.Uint:
		if tag.DefaultVal.Value == nil {
			c.flags.Uint(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint()
		if !ok {
			return fmt.Errorf("failed to unwrap uint")
		}
		c.flags.Uint(tag.Name, v, tag.Description)
	case reflect.Float32:
		if tag.DefaultVal.Value == nil {
			c.flags.Float32(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapFloat32()
		if !ok {
			return fmt.Errorf("failed to unwrap float32")
		}
		c.flags.Float32(tag.Name, v, tag.Description)
	case reflect.Float64:
		if tag.DefaultVal.Value == nil {
			c.flags.Float64(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapFloat64()
		if !ok {
			return fmt.Errorf("failed to unwrap float64")
		}
		c.flags.Float64(tag.Name, v, tag.Description)
	case reflect.Complex64, reflect.Complex128:
		return fmt.Errorf("complex types are not supported in pflags")
	case reflect.String:
		if tag.DefaultVal.Value == nil {
			c.flags.String(tag.Name, "", tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapString()
		if !ok {
			return fmt.Errorf("failed to unwrap string")
		}
		c.flags.String(tag.Name, v, tag.Description)
	case reflect.Struct:
		sep := "."
		if c.pflagOptions != nil && c.pflagOptions.Separator != "" {
			sep = c.pflagOptions.Separator
		}
		err = c.registerFlagsFromStruct(reflect.New(field.Type).Interface(), tag.Name+sep)
		if err != nil {
			return fmt.Errorf("failed to register flags for struct %s: %w", field.Name, err)
		}
	case reflect.Array, reflect.Slice:
		switch field.Type.Elem().Kind() {
		case reflect.Bool:
			if tag.DefaultVal.Value == nil {
				c.flags.BoolSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapBoolSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap bool slice")
			}
			c.flags.BoolSlice(tag.Name, v, tag.Description)
		case reflect.Float32:
			if tag.DefaultVal.Value == nil {
				c.flags.Float32Slice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapFloat32Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap float32 slice")
			}
			c.flags.Float32Slice(tag.Name, v, tag.Description)
		case reflect.Float64:
			if tag.DefaultVal.Value == nil {
				c.flags.Float64Slice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapFloat64Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap float64 slice")
			}
			c.flags.Float64Slice(tag.Name, v, tag.Description)
		case reflect.Int32:
			if tag.DefaultVal.Value == nil {
				c.flags.Int32Slice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapInt32Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap int32 slice")
			}
			c.flags.Int32Slice(tag.Name, v, tag.Description)
		case reflect.Int64:
			if tag.DefaultVal.Value == nil {
				c.flags.Int64Slice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapInt64Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap int64 slice")
			}
			c.flags.Int64Slice(tag.Name, v, tag.Description)
		case reflect.Int:
			if tag.DefaultVal.Value == nil {
				c.flags.IntSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapIntSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap int slice")
			}
			c.flags.IntSlice(tag.Name, v, tag.Description)
		case reflect.String:
			if tag.DefaultVal.Value == nil {
				c.flags.StringSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapStringSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap string slice")
			}
			c.flags.StringSlice(tag.Name, v, tag.Description)
		case reflect.Uint:
			if tag.DefaultVal.Value == nil {
				c.flags.UintSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapUintSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap uint slice")
			}
			c.flags.UintSlice(tag.Name, v, tag.Description)
		case reflect.Uint8:
			if tag.DefaultVal.Value == nil {
				c.flags.UintSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapUint8Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap uint8 slice")
			}
			uint8Slice := make([]uint, len(v))
			for i, val := range v {
				uint8Slice[i] = uint(val)
			}
			c.flags.UintSlice(tag.Name, uint8Slice, tag.Description)
		case reflect.Uint16:
			if tag.DefaultVal.Value == nil {
				c.flags.UintSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapUint16Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap uint16 slice")
			}
			uint16Slice := make([]uint, len(v))
			for i, val := range v {
				uint16Slice[i] = uint(val)
			}
			c.flags.UintSlice(tag.Name, uint16Slice, tag.Description)
		case reflect.Uint32:
			if tag.DefaultVal.Value == nil {
				c.flags.UintSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapUint32Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap uint32 slice")
			}
			uint32Slice := make([]uint, len(v))
			for i, val := range v {
				uint32Slice[i] = uint(val)
			}
			c.flags.UintSlice(tag.Name, uint32Slice, tag.Description)
		case reflect.Uint64:
			if tag.DefaultVal.Value == nil {
				c.flags.UintSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapUint64Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap uint64 slice")
			}
			uint64Slice := make([]uint, len(v))
			for i, val := range v {
				uint64Slice[i] = uint(val)
			}
			c.flags.UintSlice(tag.Name, uint64Slice, tag.Description)
		case reflect.Int8:
			if tag.DefaultVal.Value == nil {
				c.flags.IntSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapInt8Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap int8 slice")
			}
			int8Slice := make([]int, len(v))
			for i, val := range v {
				int8Slice[i] = int(val)
			}
			c.flags.IntSlice(tag.Name, int8Slice, tag.Description)
		case reflect.Int16:
			if tag.DefaultVal.Value == nil {
				c.flags.IntSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapInt16Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap int16 slice")
			}
			int16Slice := make([]int, len(v))
			for i, val := range v {
				int16Slice[i] = int(val)
			}
			c.flags.IntSlice(tag.Name, int16Slice, tag.Description)
		case reflect.Complex64, reflect.Complex128:
			return fmt.Errorf("complex types are not supported")
		case reflect.Invalid:
			return fmt.Errorf("invalid type in config: %v", field.Type)
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			return fmt.Errorf("unsupported type in config: %v", field.Type)
		default:
			return fmt.Errorf("unsupported type in config: %v", field.Type)
		}
	case reflect.Invalid:
		return fmt.Errorf("invalid type in config: %v", field.Type)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Errorf("unsupported type in config: %v", field.Type)
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
