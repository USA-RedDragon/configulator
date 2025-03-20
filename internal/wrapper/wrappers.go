package wrapper

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type WrappedValue struct {
	Value any
}

func (w WrappedValue) UnwrapBool() (bool, bool) {
	if w.Value == nil {
		return false, false
	}
	b, ok := w.Value.(bool)
	return b, ok
}

func (w WrappedValue) UnwrapInt8() (int8, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(int8)
	return i, ok
}

func (w WrappedValue) UnwrapInt16() (int16, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(int16)
	return i, ok
}

func (w WrappedValue) UnwrapInt32() (int32, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(int32)
	return i, ok
}

func (w WrappedValue) UnwrapInt64() (int64, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(int64)
	return i, ok
}

func (w WrappedValue) UnwrapInt() (int, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(int)
	return i, ok
}

func (w WrappedValue) UnwrapUint8() (uint8, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(uint8)
	return i, ok
}

func (w WrappedValue) UnwrapUint16() (uint16, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(uint16)
	return i, ok
}

func (w WrappedValue) UnwrapUint32() (uint32, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(uint32)
	return i, ok
}

func (w WrappedValue) UnwrapUint64() (uint64, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(uint64)
	return i, ok
}

func (w WrappedValue) UnwrapUint() (uint, bool) {
	if w.Value == nil {
		return 0, false
	}
	i, ok := w.Value.(uint)
	return i, ok
}

func (w WrappedValue) UnwrapString() (string, bool) {
	if w.Value == nil {
		return "", false
	}
	s, ok := w.Value.(string)
	return s, ok
}

func (w WrappedValue) UnwrapFloat32() (float32, bool) {
	if w.Value == nil {
		return 0, false
	}
	f, ok := w.Value.(float32)
	return f, ok
}

func (w WrappedValue) UnwrapFloat64() (float64, bool) {
	if w.Value == nil {
		return 0, false
	}
	f, ok := w.Value.(float64)
	return f, ok
}

func (w WrappedValue) UnwrapBoolSlice() ([]bool, bool) {
	if w.Value == nil {
		return nil, false
	}
	b, ok := w.Value.([]bool)
	return b, ok
}

func (w WrappedValue) UnwrapFloat32Slice() ([]float32, bool) {
	if w.Value == nil {
		return nil, false
	}
	f, ok := w.Value.([]float32)
	return f, ok
}

func (w WrappedValue) UnwrapFloat64Slice() ([]float64, bool) {
	if w.Value == nil {
		return nil, false
	}
	f, ok := w.Value.([]float64)
	return f, ok
}

func (w WrappedValue) UnwrapInt8Slice() ([]int8, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]int8)
	return i, ok
}

func (w WrappedValue) UnwrapInt16Slice() ([]int16, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]int16)
	return i, ok
}

func (w WrappedValue) UnwrapInt32Slice() ([]int32, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]int32)
	return i, ok
}

func (w WrappedValue) UnwrapInt64Slice() ([]int64, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]int64)
	return i, ok
}

func (w WrappedValue) UnwrapIntSlice() ([]int, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]int)
	return i, ok
}

func (w WrappedValue) UnwrapStringSlice() ([]string, bool) {
	if w.Value == nil {
		return nil, false
	}
	s, ok := w.Value.([]string)
	return s, ok
}

func (w WrappedValue) UnwrapUint8Slice() ([]uint8, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]uint8)
	return i, ok
}

func (w WrappedValue) UnwrapUint16Slice() ([]uint16, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]uint16)
	return i, ok
}

func (w WrappedValue) UnwrapUint32Slice() ([]uint32, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]uint32)
	return i, ok
}

func (w WrappedValue) UnwrapUint64Slice() ([]uint64, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]uint64)
	return i, ok
}

func (w WrappedValue) UnwrapUintSlice() ([]uint, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]uint)
	return i, ok
}

func (w WrappedValue) UnwrapInterface() (any, bool) {
	if w.Value == nil {
		return nil, false
	}
	return w.Value, true
}

func (w WrappedValue) UnwrapInterfaceSlice() ([]any, bool) {
	if w.Value == nil {
		return nil, false
	}
	i, ok := w.Value.([]any)
	return i, ok
}

func WrapString(typ reflect.Type, val, arraySeparator string) (WrappedValue, error) {
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
		case reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
			return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Elem().Kind())
		default:
			return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Elem().Kind())
		}
	case reflect.Invalid:
		return WrappedValue{}, fmt.Errorf("invalid type %v", typ.Kind())
	case reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
		return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Kind())
	default:
		return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Kind())
	}
}
