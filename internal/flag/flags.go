package flag

import (
	"fmt"
	"reflect"

	"github.com/USA-RedDragon/configulator/internal/tags"
	"github.com/spf13/pflag"
)

func RegisterFlagsFromStruct(flags *pflag.FlagSet, stru any, prefix, separator, arraySeparator string) error {
	typ := reflect.TypeOf(stru).Elem()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", typ.Kind())
	}

	for i := range typ.NumField() {
		field := typ.Field(i)
		if tag := field.Tag.Get("name"); tag != "" {
			err := AddFlag(flags, prefix, field, separator, arraySeparator)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//nolint:golint,gocyclo
func AddFlag(flags *pflag.FlagSet, prefix string, field reflect.StructField, separator, arraySeparator string) error {
	tag, err := tags.ExtractStructTags(field, arraySeparator)
	if err != nil {
		return err
	}

	tag.Name = prefix + tag.Name

	switch field.Type.Kind() {
	case reflect.Bool:
		if tag.DefaultVal.Value == nil {
			flags.Bool(tag.Name, false, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapBool()
		if !ok {
			return fmt.Errorf("failed to unwrap bool")
		}
		flags.Bool(tag.Name, v, tag.Description)
	case reflect.Int8:
		if tag.DefaultVal.Value == nil {
			flags.Int8(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt8()
		if !ok {
			return fmt.Errorf("failed to unwrap int8")
		}
		flags.Int8(tag.Name, v, tag.Description)
	case reflect.Int16:
		if tag.DefaultVal.Value == nil {
			flags.Int16(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt16()
		if !ok {
			return fmt.Errorf("failed to unwrap int16")
		}
		flags.Int16(tag.Name, v, tag.Description)
	case reflect.Int32:
		if tag.DefaultVal.Value == nil {
			flags.Int32(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt32()
		if !ok {
			return fmt.Errorf("failed to unwrap int32")
		}
		flags.Int32(tag.Name, v, tag.Description)
	case reflect.Int64:
		if tag.DefaultVal.Value == nil {
			flags.Int64(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt64()
		if !ok {
			return fmt.Errorf("failed to unwrap int64")
		}
		flags.Int64(tag.Name, v, tag.Description)
	case reflect.Int:
		if tag.DefaultVal.Value == nil {
			flags.Int(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInt()
		if !ok {
			return fmt.Errorf("failed to unwrap int")
		}
		flags.Int(tag.Name, v, tag.Description)
	case reflect.Uint8:
		if tag.DefaultVal.Value == nil {
			flags.Uint8(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint8()
		if !ok {
			return fmt.Errorf("failed to unwrap uint8")
		}
		flags.Uint8(tag.Name, v, tag.Description)
	case reflect.Uint16:
		if tag.DefaultVal.Value == nil {
			flags.Uint16(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint16()
		if !ok {
			return fmt.Errorf("failed to unwrap uint16")
		}
		flags.Uint16(tag.Name, v, tag.Description)
	case reflect.Uint32:
		if tag.DefaultVal.Value == nil {
			flags.Uint32(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint32()
		if !ok {
			return fmt.Errorf("failed to unwrap uint32")
		}
		flags.Uint32(tag.Name, v, tag.Description)
	case reflect.Uint64:
		if tag.DefaultVal.Value == nil {
			flags.Uint64(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint64()
		if !ok {
			return fmt.Errorf("failed to unwrap uint64")
		}
		flags.Uint64(tag.Name, v, tag.Description)
	case reflect.Uint:
		if tag.DefaultVal.Value == nil {
			flags.Uint(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapUint()
		if !ok {
			return fmt.Errorf("failed to unwrap uint")
		}
		flags.Uint(tag.Name, v, tag.Description)
	case reflect.Float32:
		if tag.DefaultVal.Value == nil {
			flags.Float32(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapFloat32()
		if !ok {
			return fmt.Errorf("failed to unwrap float32")
		}
		flags.Float32(tag.Name, v, tag.Description)
	case reflect.Float64:
		if tag.DefaultVal.Value == nil {
			flags.Float64(tag.Name, 0, tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapFloat64()
		if !ok {
			return fmt.Errorf("failed to unwrap float64")
		}
		flags.Float64(tag.Name, v, tag.Description)
	case reflect.Complex64, reflect.Complex128:
		return fmt.Errorf("complex types are not supported in pflags")
	case reflect.String:
		if tag.DefaultVal.Value == nil {
			flags.String(tag.Name, "", tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapString()
		if !ok {
			return fmt.Errorf("failed to unwrap string")
		}
		flags.String(tag.Name, v, tag.Description)
	case reflect.Interface:
		if tag.DefaultVal.Value == nil {
			flags.String(tag.Name, "", tag.Description)
			return nil
		}
		v, ok := tag.DefaultVal.UnwrapInterface()
		if !ok {
			return fmt.Errorf("failed to unwrap string")
		}
		str := fmt.Sprintf("%v", v)
		flags.String(tag.Name, str, tag.Description)
	case reflect.Struct:
		err = RegisterFlagsFromStruct(flags, reflect.New(field.Type).Interface(), tag.Name+separator, separator, arraySeparator)
		if err != nil {
			return fmt.Errorf("failed to register flags for struct %s: %w", field.Name, err)
		}
	case reflect.Array, reflect.Slice:
		switch field.Type.Elem().Kind() {
		case reflect.Bool:
			if tag.DefaultVal.Value == nil {
				flags.BoolSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapBoolSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap bool slice")
			}
			flags.BoolSlice(tag.Name, v, tag.Description)
		case reflect.Float32:
			if tag.DefaultVal.Value == nil {
				flags.Float32Slice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapFloat32Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap float32 slice")
			}
			flags.Float32Slice(tag.Name, v, tag.Description)
		case reflect.Float64:
			if tag.DefaultVal.Value == nil {
				flags.Float64Slice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapFloat64Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap float64 slice")
			}
			flags.Float64Slice(tag.Name, v, tag.Description)
		case reflect.Int32:
			if tag.DefaultVal.Value == nil {
				flags.Int32Slice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapInt32Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap int32 slice")
			}
			flags.Int32Slice(tag.Name, v, tag.Description)
		case reflect.Int64:
			if tag.DefaultVal.Value == nil {
				flags.Int64Slice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapInt64Slice()
			if !ok {
				return fmt.Errorf("failed to unwrap int64 slice")
			}
			flags.Int64Slice(tag.Name, v, tag.Description)
		case reflect.Int:
			if tag.DefaultVal.Value == nil {
				flags.IntSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapIntSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap int slice")
			}
			flags.IntSlice(tag.Name, v, tag.Description)
		case reflect.String:
			if tag.DefaultVal.Value == nil {
				flags.StringSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapStringSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap string slice")
			}
			flags.StringSlice(tag.Name, v, tag.Description)
		case reflect.Uint:
			if tag.DefaultVal.Value == nil {
				flags.UintSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapUintSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap uint slice")
			}
			flags.UintSlice(tag.Name, v, tag.Description)
		case reflect.Uint8:
			if tag.DefaultVal.Value == nil {
				flags.UintSlice(tag.Name, nil, tag.Description)
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
			flags.UintSlice(tag.Name, uint8Slice, tag.Description)
		case reflect.Uint16:
			if tag.DefaultVal.Value == nil {
				flags.UintSlice(tag.Name, nil, tag.Description)
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
			flags.UintSlice(tag.Name, uint16Slice, tag.Description)
		case reflect.Uint32:
			if tag.DefaultVal.Value == nil {
				flags.UintSlice(tag.Name, nil, tag.Description)
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
			flags.UintSlice(tag.Name, uint32Slice, tag.Description)
		case reflect.Uint64:
			if tag.DefaultVal.Value == nil {
				flags.UintSlice(tag.Name, nil, tag.Description)
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
			flags.UintSlice(tag.Name, uint64Slice, tag.Description)
		case reflect.Int8:
			if tag.DefaultVal.Value == nil {
				flags.IntSlice(tag.Name, nil, tag.Description)
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
			flags.IntSlice(tag.Name, int8Slice, tag.Description)
		case reflect.Int16:
			if tag.DefaultVal.Value == nil {
				flags.IntSlice(tag.Name, nil, tag.Description)
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
			flags.IntSlice(tag.Name, int16Slice, tag.Description)
		case reflect.Interface:
			if tag.DefaultVal.Value == nil {
				flags.StringSlice(tag.Name, nil, tag.Description)
				return nil
			}
			v, ok := tag.DefaultVal.UnwrapInterfaceSlice()
			if !ok {
				return fmt.Errorf("failed to unwrap string slice")
			}
			strSlice := make([]string, len(v))
			for i, val := range v {
				strSlice[i] = fmt.Sprintf("%v", val)
			}
			flags.StringSlice(tag.Name, strSlice, tag.Description)
		case reflect.Complex64, reflect.Complex128:
			return fmt.Errorf("complex types are not supported")
		case reflect.Invalid:
			return fmt.Errorf("invalid type in config: %v", field.Type)
		case reflect.Pointer, reflect.Uintptr:
			return fmt.Errorf("pointer types are not supported")
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			return fmt.Errorf("unsupported type in config: %v", field.Type)
		default:
			return fmt.Errorf("unsupported type in config: %v", field.Type)
		}
	case reflect.Invalid:
		return fmt.Errorf("invalid type in config: %v", field.Type)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Errorf("unsupported type in config: %v", field.Type)
	case reflect.Pointer, reflect.Uintptr:
		return fmt.Errorf("pointer types are not supported")
	default:
		return fmt.Errorf("unsupported type in config: %v", field.Type)
	}
	return nil
}
