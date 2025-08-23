package wrapper

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

//nolint:golint,gocyclo
func WrapString(typ reflect.Type, val, arraySeparator string) (WrappedValue, error) {
	if (typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array) && strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
		val = strings.Trim(val, "[]")
	}
	switch typ.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{b}, nil
	case reflect.Int8:
		i, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{int8(i)}, nil
	case reflect.Int16:
		i, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{int16(i)}, nil
	case reflect.Int32:
		i, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{int32(i)}, nil
	case reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{i}, nil
	case reflect.Int:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{int(i)}, nil
	case reflect.Uint8:
		i, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{uint8(i)}, nil
	case reflect.Uint16:
		i, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{uint16(i)}, nil
	case reflect.Uint32:
		i, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{uint32(i)}, nil
	case reflect.Uint64:
		i, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{i}, nil
	case reflect.Uint:
		i, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{uint(i)}, nil
	case reflect.String:
		return WrappedValue{val}, nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{float32(f)}, nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return WrappedValue{}, err
		}
		return WrappedValue{f}, nil
	case reflect.Interface:
		return WrappedValue{val}, nil
	case reflect.Slice, reflect.Array:
		switch typ.Elem().Kind() {
		case reflect.Bool:
			parts := strings.Split(val, arraySeparator)
			bools := make([]bool, len(parts))
			for i, part := range parts {
				b, err := strconv.ParseBool(part)
				if err != nil {
					return WrappedValue{}, err
				}
				bools[i] = b
			}
			return WrappedValue{bools}, nil
		case reflect.Float32:
			parts := strings.Split(val, arraySeparator)
			floats := make([]float32, len(parts))
			for i, part := range parts {
				f, err := strconv.ParseFloat(part, 32)
				if err != nil {
					return WrappedValue{}, err
				}
				floats[i] = float32(f)
			}
			return WrappedValue{floats}, nil
		case reflect.Float64:
			parts := strings.Split(val, arraySeparator)
			floats := make([]float64, len(parts))
			for i, part := range parts {
				f, err := strconv.ParseFloat(part, 64)
				if err != nil {
					return WrappedValue{}, err
				}
				floats[i] = f
			}
			return WrappedValue{floats}, nil
		case reflect.Int8:
			parts := strings.Split(val, arraySeparator)
			ints := make([]int8, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseInt(part, 10, 8)
				if err != nil {
					return WrappedValue{}, err
				}
				ints[i] = int8(in)
			}
			return WrappedValue{ints}, nil
		case reflect.Int16:
			parts := strings.Split(val, arraySeparator)
			ints := make([]int16, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseInt(part, 10, 16)
				if err != nil {
					return WrappedValue{}, err
				}
				ints[i] = int16(in)
			}
			return WrappedValue{ints}, nil
		case reflect.Int32:
			parts := strings.Split(val, arraySeparator)
			ints := make([]int32, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseInt(part, 10, 32)
				if err != nil {
					return WrappedValue{}, err
				}
				ints[i] = int32(in)
			}
			return WrappedValue{ints}, nil
		case reflect.Int64:
			parts := strings.Split(val, arraySeparator)
			ints := make([]int64, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return WrappedValue{}, err
				}
				ints[i] = in
			}
			return WrappedValue{ints}, nil
		case reflect.Int:
			parts := strings.Split(val, arraySeparator)
			ints := make([]int, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return WrappedValue{}, err
				}
				ints[i] = int(in)
			}
			return WrappedValue{ints}, nil
		case reflect.String:
			return WrappedValue{strings.Split(val, arraySeparator)}, nil
		case reflect.Uint8:
			parts := strings.Split(val, arraySeparator)
			uints := make([]uint8, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseUint(part, 10, 8)
				if err != nil {
					return WrappedValue{}, err
				}
				uints[i] = uint8(in)
			}
			return WrappedValue{uints}, nil
		case reflect.Uint16:
			parts := strings.Split(val, arraySeparator)
			uints := make([]uint16, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseUint(part, 10, 16)
				if err != nil {
					return WrappedValue{}, err
				}
				uints[i] = uint16(in)
			}
			return WrappedValue{uints}, nil
		case reflect.Uint32:
			parts := strings.Split(val, arraySeparator)
			uints := make([]uint32, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseUint(part, 10, 32)
				if err != nil {
					return WrappedValue{}, err
				}
				uints[i] = uint32(in)
			}
			return WrappedValue{uints}, nil
		case reflect.Uint64:
			parts := strings.Split(val, arraySeparator)
			uints := make([]uint64, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseUint(part, 10, 64)
				if err != nil {
					return WrappedValue{}, err
				}
				uints[i] = in
			}
			return WrappedValue{uints}, nil
		case reflect.Uint:
			parts := strings.Split(val, arraySeparator)
			uints := make([]uint, len(parts))
			for i, part := range parts {
				in, err := strconv.ParseUint(part, 10, 64)
				if err != nil {
					return WrappedValue{}, err
				}
				uints[i] = uint(in)
			}
			return WrappedValue{uints}, nil
		case reflect.Interface:
			parts := strings.Split(val, arraySeparator)
			interfaces := make([]any, len(parts))
			for i, part := range parts {
				interfaces[i] = part
			}
			return WrappedValue{interfaces}, nil
		case reflect.Invalid:
			return WrappedValue{}, fmt.Errorf("invalid type %v", typ.Elem().Kind())
		case reflect.Complex64, reflect.Complex128:
			return WrappedValue{}, fmt.Errorf("complex types are not supported")
		case reflect.Pointer, reflect.Uintptr:
			return WrappedValue{}, fmt.Errorf("pointer types are not supported")
		case reflect.Struct:
			// a struct itself can't be expressed in an environment variable, so we'll never get here
			return WrappedValue{}, fmt.Errorf("unsupported struct type %v", typ.Elem().Kind())
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Elem().Kind())
		default:
			return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Elem().Kind())
		}
	case reflect.Pointer, reflect.Uintptr:
		return WrappedValue{}, fmt.Errorf("pointer types are not supported")
	case reflect.Complex64, reflect.Complex128:
		return WrappedValue{}, fmt.Errorf("complex types are not supported")
	case reflect.Invalid:
		return WrappedValue{}, fmt.Errorf("invalid type %v", typ.Kind())
	case reflect.Struct:
		// a struct itself can't be expressed in an environment variable, so we'll never get here
		return WrappedValue{}, fmt.Errorf("unsupported struct type %v", typ.Kind())
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Kind())
	default:
		return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Kind())
	}
}
