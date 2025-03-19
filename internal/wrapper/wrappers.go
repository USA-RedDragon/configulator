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

func (w WrappedValue) UnwrapBool() bool {
	if w.Value == nil {
		return false
	}
	return w.Value.(bool)
}

func (w WrappedValue) UnwrapInt8() int8 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(int8)
}

func (w WrappedValue) UnwrapInt16() int16 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(int16)
}

func (w WrappedValue) UnwrapInt32() int32 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(int32)
}

func (w WrappedValue) UnwrapInt64() int64 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(int64)
}

func (w WrappedValue) UnwrapInt() int {
	if w.Value == nil {
		return 0
	}
	return w.Value.(int)
}

func (w WrappedValue) UnwrapUint8() uint8 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(uint8)
}

func (w WrappedValue) UnwrapUint16() uint16 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(uint16)
}

func (w WrappedValue) UnwrapUint32() uint32 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(uint32)
}

func (w WrappedValue) UnwrapUint64() uint64 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(uint64)
}

func (w WrappedValue) UnwrapUint() uint {
	if w.Value == nil {
		return 0
	}
	return w.Value.(uint)
}

func (w WrappedValue) UnwrapString() string {
	if w.Value == nil {
		return ""
	}
	return w.Value.(string)
}

func (w WrappedValue) UnwrapFloat32() float32 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(float32)
}

func (w WrappedValue) UnwrapFloat64() float64 {
	if w.Value == nil {
		return 0
	}
	return w.Value.(float64)
}

func (w WrappedValue) UnwrapBoolSlice() []bool {
	if w.Value == nil {
		return []bool{}
	}
	return w.Value.([]bool)
}

func (w WrappedValue) UnwrapFloat32Slice() []float32 {
	if w.Value == nil {
		return []float32{}
	}
	return w.Value.([]float32)
}

func (w WrappedValue) UnwrapFloat64Slice() []float64 {
	if w.Value == nil {
		return []float64{}
	}
	return w.Value.([]float64)
}

func (w WrappedValue) UnwrapInt32Slice() []int32 {
	if w.Value == nil {
		return []int32{}
	}
	return w.Value.([]int32)
}

func (w WrappedValue) UnwrapInt64Slice() []int64 {
	if w.Value == nil {
		return []int64{}
	}
	return w.Value.([]int64)
}

func (w WrappedValue) UnwrapIntSlice() []int {
	if w.Value == nil {
		return []int{}
	}
	return w.Value.([]int)
}

func (w WrappedValue) UnwrapStringSlice() []string {
	if w.Value == nil {
		return []string{}
	}
	return w.Value.([]string)
}

func (w WrappedValue) UnwrapUintSlice() []uint {
	if w.Value == nil {
		return []uint{}
	}
	return w.Value.([]uint)
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
		default:
			return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Elem().Kind())
		}
	default:
		return WrappedValue{}, fmt.Errorf("unsupported type %v", typ.Kind())
	}
}
