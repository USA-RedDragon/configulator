package reflect

import (
	"fmt"
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
		if tag := field.Tag.Get("config"); tag != "" {
			tagInfo, err := tags.ExtractStructTag(field, tag, arraySeparator)
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

func SetStructValue(stru *reflect.Value, field reflect.StructField, val wrapper.WrappedValue) error {
	v := stru.FieldByName(field.Name)
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(val.UnwrapBool())
	case reflect.Int8:
		v.SetInt(int64(val.UnwrapInt8()))
	case reflect.Int16:
		v.SetInt(int64(val.UnwrapInt16()))
	case reflect.Int32:
		v.SetInt(int64(val.UnwrapInt32()))
	case reflect.Int64:
		v.SetInt(val.UnwrapInt64())
	case reflect.Int:
		v.SetInt(int64(val.UnwrapInt()))
	case reflect.Uint8:
		v.SetUint(uint64(val.UnwrapUint8()))
	case reflect.Uint16:
		v.SetUint(uint64(val.UnwrapUint16()))
	case reflect.Uint32:
		v.SetUint(uint64(val.UnwrapUint32()))
	case reflect.Uint64:
		v.SetUint(val.UnwrapUint64())
	case reflect.Uint:
		v.SetUint(uint64(val.UnwrapUint()))
	case reflect.Float32:
		v.SetFloat(float64(val.UnwrapFloat32()))
	case reflect.Float64:
		v.SetFloat(val.UnwrapFloat64())
	case reflect.String:
		v.SetString(val.UnwrapString())
	case reflect.Struct:
		err := SetStructValue(&v, field, val)
		if err != nil {
			return err
		}
	case reflect.Array, reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Bool:
			v.Set(reflect.ValueOf(val.UnwrapBoolSlice()))
		case reflect.Float32:
			v.Set(reflect.ValueOf(val.UnwrapFloat32Slice()))
		case reflect.Float64:
			v.Set(reflect.ValueOf(val.UnwrapFloat64Slice()))
		case reflect.Int32:
			v.Set(reflect.ValueOf(val.UnwrapInt32Slice()))
		case reflect.Int64:
			v.Set(reflect.ValueOf(val.UnwrapInt64Slice()))
		case reflect.Int:
			v.Set(reflect.ValueOf(val.UnwrapIntSlice()))
		case reflect.String:
			v.Set(reflect.ValueOf(val.UnwrapStringSlice()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			v.Set(reflect.ValueOf(val.UnwrapUintSlice()))
		default:
			return fmt.Errorf("unsupported array/slice type in config: %v", field.Type)
		}
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
		if tag := f.Tag.Get("config"); tag != "" {
			tagInfo, err := tags.ExtractStructTag(f, tag, arraySeparator)
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

func SetNestedStructValue(val *reflect.Value, typ reflect.Type, fields []string, wrapped wrapper.WrappedValue, arraySeparator string) error {
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
