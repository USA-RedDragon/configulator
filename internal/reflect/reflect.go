package reflect

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/USA-RedDragon/configulator/internal/tags"
	"github.com/USA-RedDragon/configulator/internal/wrapper"
)

func GetDefaultsFromStruct(typ reflect.Type, arraySeparator string) (any, error) {
	var ret any
	if typ.Kind() != reflect.Struct {
		return ret, fmt.Errorf("expected struct, got %v", typ.Kind())
	}

	newVal := reflect.New(typ).Elem()

	for i := range typ.NumField() {
		field := typ.Field(i)
		if tag := field.Tag.Get("name"); tag != "" {
			tagInfo, err := tags.ExtractStructTags(field, arraySeparator)
			if err != nil {
				return ret, err
			}
			structField := newVal.FieldByName(field.Name)
			if structField.Kind() == reflect.Struct {
				nested, err := GetDefaultsFromStruct(structField.Type(), arraySeparator)
				if err != nil {
					return ret, err
				}
				structField.Set(reflect.ValueOf(nested))
				continue
			}
			err = SetStructValue(&newVal, field, tagInfo.DefaultVal)
			if err != nil {
				return ret, err
			}
		}
	}

	ret = newVal.Interface()

	return ret, nil
}

//nolint:golint,gocyclo
func SetStructValue(stru *reflect.Value, field reflect.StructField, val wrapper.WrappedValue) error {
	v := stru.FieldByName(field.Name)
	switch v.Kind() {
	case reflect.Bool:
		var b bool
		if val.Value != nil {
			var ok bool
			b, ok = val.UnwrapBool()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if b, ok = a.(bool); ok {
						v.SetBool(b)
						return nil
					}
				}
				return fmt.Errorf("failed to unwrap bool")
			}
		}
		v.SetBool(b)
	case reflect.Int8:
		var i int8
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapInt8()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(int8); ok {
						v.SetInt(int64(i))
						return nil
					}
				}
				return fmt.Errorf("failed to unwrap int8")
			}
		}
		v.SetInt(int64(i))
	case reflect.Int16:
		var i int16
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapInt16()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(int16); ok {
						v.SetInt(int64(i))
						return nil
					}
				}
				return fmt.Errorf("failed to unwrap int16")
			}
		}
		v.SetInt(int64(i))
	case reflect.Int32:
		var i int32
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapInt32()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(int32); ok {
						v.SetInt(int64(i))
						return nil
					}
				}
				return fmt.Errorf("failed to unwrap int32")
			}
		}
		v.SetInt(int64(i))
	case reflect.Int64:
		var i int64
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapInt64()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(int64); ok {
						v.SetInt(i)
						return nil
					}
				}
				return fmt.Errorf("failed to unwrap int64")
			}
		}
		v.SetInt(i)
	case reflect.Int:
		var i int
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapInt()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(int); ok {
						v.SetInt(int64(i))
						return nil
					}
				}
				return fmt.Errorf("failed to unwrap int")
			}
		}
		v.SetInt(int64(i))
	case reflect.Uint8:
		var i uint8
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapUint8()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(uint8); ok {
						v.SetUint(uint64(i))
						return nil
					}
				}
				return fmt.Errorf("failed to unwrap uint8")
			}
		}
		v.SetUint(uint64(i))
	case reflect.Uint16:
		var i uint16
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapUint16()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(uint16); ok {
						v.SetUint(uint64(i))
						return nil
					} else {
						return fmt.Errorf("failed to unwrap uint16")
					}
				} else {
					return fmt.Errorf("failed to unwrap uint16")
				}
			}
		}
		v.SetUint(uint64(i))
	case reflect.Uint32:
		var i uint32
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapUint32()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(uint32); ok {
						v.SetUint(uint64(i))
						return nil
					} else {
						return fmt.Errorf("failed to unwrap uint32")
					}
				} else {
					return fmt.Errorf("failed to unwrap uint32")
				}
			}
		}
		v.SetUint(uint64(i))
	case reflect.Uint64:
		var i uint64
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapUint64()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(uint64); ok {
						v.SetUint(i)
						return nil
					} else {
						return fmt.Errorf("failed to unwrap uint64")
					}
				} else {
					return fmt.Errorf("failed to unwrap uint64")
				}
			}
		}
		v.SetUint(i)
	case reflect.Uint:
		var i uint
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapUint()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if i, ok = a.(uint); ok {
						v.SetUint(uint64(i))
						return nil
					} else {
						return fmt.Errorf("failed to unwrap uint")
					}
				} else {
					return fmt.Errorf("failed to unwrap uint")
				}
			}
		}
		v.SetUint(uint64(i))
	case reflect.Float32:
		var f float32
		if val.Value != nil {
			var ok bool
			f, ok = val.UnwrapFloat32()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if f, ok = a.(float32); ok {
						v.SetFloat(float64(f))
						return nil
					} else {
						return fmt.Errorf("failed to unwrap float32")
					}
				} else {
					return fmt.Errorf("failed to unwrap float32")
				}
			}
		}
		v.SetFloat(float64(f))
	case reflect.Float64:
		var f float64
		if val.Value != nil {
			var ok bool
			f, ok = val.UnwrapFloat64()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if f, ok = a.(float64); ok {
						v.SetFloat(f)
						return nil
					} else {
						return fmt.Errorf("failed to unwrap float64")
					}
				} else {
					return fmt.Errorf("failed to unwrap float64")
				}
			}
		}
		v.SetFloat(f)
	case reflect.Interface:
		var i any
		if val.Value != nil {
			var ok bool
			i, ok = val.UnwrapInterface()
			if !ok {
				return fmt.Errorf("failed to unwrap interface")
			}
		}
		if i != nil {
			v.Set(reflect.ValueOf(i))
		}
	case reflect.String:
		var s string
		if val.Value != nil {
			var ok bool
			s, ok = val.UnwrapString()
			if !ok {
				a, ok := val.UnwrapInterface()
				if ok {
					if s, ok = a.(string); ok {
						v.SetString(s)
						return nil
					} else {
						return fmt.Errorf("failed to unwrap string")
					}
				} else {
					return fmt.Errorf("failed to unwrap string")
				}
			}
		}
		v.SetString(s)
	case reflect.Struct:
		err := SetStructValue(&v, field, val)
		if err != nil {
			return err
		}
	case reflect.Array, reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Bool:
			var s []bool
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapBoolSlice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]bool); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap bool slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap bool slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Float32:
			var s []float32
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapFloat32Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]float32); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap float32 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap float32 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Float64:
			var s []float64
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapFloat64Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]float64); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap float64 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap float64 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Int8:
			var s []int8
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapInt8Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]int8); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap int8 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap int8 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Int16:
			var s []int16
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapInt16Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]int16); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap int16 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap int16 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Int32:
			var s []int32
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapInt32Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]int32); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap int32 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap int32 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Int64:
			var s []int64
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapInt64Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]int64); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap int64 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap int64 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Int:
			var s []int
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapIntSlice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]int); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap int slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap int slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Interface:
			var s []any
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapInterfaceSlice()
				if !ok {
					return fmt.Errorf("failed to unwrap interface slice")
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.String:
			var s []string
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapStringSlice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						slog.Info("Setting string slice", "val", s, "type", reflect.TypeOf(s))
						if s, ok = s.([]string); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap string slice %v=%v", reflect.TypeOf(s), s)
						}
					} else {
						return fmt.Errorf("failed to unwrap string slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Uint8:
			var s []uint8
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapUint8Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]uint8); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap uint8 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap uint8 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Uint16:
			var s []uint16
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapUint16Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]uint16); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap uint16 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap uint16 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Uint32:
			var s []uint32
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapUint32Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]uint32); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap uint32 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap uint32 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Uint64:
			var s []uint64
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapUint64Slice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]uint64); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap uint64 slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap uint64 slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Uint:
			var s []uint
			if val.Value != nil {
				var ok bool
				s, ok = val.UnwrapUintSlice()
				if !ok {
					s, ok := val.UnwrapInterface()
					if ok {
						if s, ok = s.([]uint); ok {
							v.Set(reflect.ValueOf(s))
							return nil
						} else {
							return fmt.Errorf("failed to unwrap uint slice")
						}
					} else {
						return fmt.Errorf("failed to unwrap uint slice")
					}
				}
			}
			v.Set(reflect.ValueOf(s))
		case reflect.Struct:
			// a struct itself can't be expressed in an environment variable, so we'll never get here
			return fmt.Errorf("unsupported struct type in config: %v", field.Type)
		case reflect.Invalid:
			return fmt.Errorf("invalid field type in config: %v", field.Type)
		case reflect.Complex64, reflect.Complex128:
			return fmt.Errorf("complex types are not supported")
		case reflect.Pointer, reflect.Uintptr:
			return fmt.Errorf("pointer types are not supported")
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			return fmt.Errorf("unsupported field type in config: %v", field.Type)
		default:
			return fmt.Errorf("unsupported array/slice type in config: %v", field.Type)
		}
	case reflect.Invalid:
		return fmt.Errorf("invalid field type in config: %v", field.Type)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Errorf("unsupported field type in config: %v", field.Type)
	case reflect.Complex64, reflect.Complex128:
		return fmt.Errorf("complex types are not supported")
	case reflect.Pointer, reflect.Uintptr:
		return fmt.Errorf("pointer types are not supported")
	default:
		return fmt.Errorf("unsupported field type in config: %v", v.Kind())
	}

	return nil
}

func GetStructFieldByName(typ reflect.Type, field, arraySeparator string) (reflect.StructField, error) {
	// searches a given struct for a `config` tag with the given field name
	// and returns the reflect.Value of the field
	if typ.Kind() != reflect.Struct {
		return reflect.StructField{}, fmt.Errorf("expected struct, got %v", typ.Kind())
	}

	for i := range typ.NumField() {
		f := typ.Field(i)
		if tag := f.Tag.Get("name"); tag != "" {
			tagInfo, err := tags.ExtractStructTags(f, arraySeparator)
			if err != nil {
				return reflect.StructField{}, err
			}
			if strings.EqualFold(tagInfo.Name, field) {
				return f, nil
			}
		}
	}

	return reflect.StructField{}, fmt.Errorf("field %s not found in struct", field)
}

func GetNestedStructFieldTypeByName(typ reflect.Type, fields []string, arraySeparator string) (*reflect.Type, error) {
	field, err := GetStructFieldByName(typ, fields[0], arraySeparator)
	if err != nil {
		return nil, fmt.Errorf("failed to get field %s: %w", fields[0], err)
	}

	// update field as we traverse the struct
	for i := 1; i < len(fields); i++ {
		field, err = GetStructFieldByName(field.Type, fields[i], arraySeparator)
		if err != nil {
			return nil, fmt.Errorf("failed to get field %s: %w", fields[i], err)
		}
	}
	return &field.Type, nil
}

func SetNestedStructValue(val *reflect.Value, fields []string, wrapped wrapper.WrappedValue, arraySeparator string) error {
	var foundFields []reflect.StructField

	field, err := GetStructFieldByName(val.Type(), fields[0], arraySeparator)
	if err != nil {
		return fmt.Errorf("failed to get field %s: %w", fields[0], err)
	}
	foundFields = append(foundFields, field)

	// update field as we traverse the struct
	for i := 1; i < len(fields); i++ {
		field, err = GetStructFieldByName(field.Type, fields[i], arraySeparator)
		if err != nil {
			return fmt.Errorf("failed to get field %s: %w", fields[i], err)
		}
		foundFields = append(foundFields, field)
	}

	// Now that we have the fields, we can set the value in the struct
	curStruct := *val
	for _, field := range foundFields {
		if curStruct.Kind() == reflect.Struct {
			nextStruct := curStruct.FieldByName(field.Name)
			if nextStruct.Kind() == reflect.Struct {
				curStruct = nextStruct
			}
		}
	}

	err = SetStructValue(&curStruct, field, wrapped)
	if err != nil {
		return fmt.Errorf("failed to set field %s value: %w", field.Name, err)
	}

	return nil
}
