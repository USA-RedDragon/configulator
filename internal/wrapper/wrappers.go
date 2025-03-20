package wrapper

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
