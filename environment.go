package configulator

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	inref "github.com/USA-RedDragon/configulator/internal/reflect"
	"github.com/USA-RedDragon/configulator/internal/tags"
	"github.com/USA-RedDragon/configulator/internal/wrapper"
)

func (c *Configulator[C]) denest(str string) []string {
	str = strings.TrimPrefix(str, c.envOptions.Prefix)
	return strings.Split(str, c.envOptions.Separator)
}

func (c *Configulator[C]) loadFromEnvironment() error {
	typ := reflect.TypeOf(c.cfg).Elem()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", typ.Kind())
	}

	envs, err := buildEnvs(typ, c.envOptions.Prefix, c.envOptions.Separator, c.arraySeparator)
	if err != nil {
		return fmt.Errorf("failed to build envs: %w", err)
	}

	for _, e := range envs {
		nested := c.denest(e.name)
		if val, exists := os.LookupEnv(e.name); exists {
			switch e.field.Type.Kind() {
			case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Float32, reflect.Float64, reflect.String, reflect.Array, reflect.Slice, reflect.Interface:
				wrapped, err := wrapper.WrapString(e.field.Type, val, c.arraySeparator)
				if err != nil {
					return fmt.Errorf("failed to wrap value: %w", err)
				}
				val := reflect.ValueOf(c.cfg).Elem()
				if err := inref.SetNestedStructValue(&val, nested, wrapped, c.arraySeparator); err != nil {
					return fmt.Errorf("failed to set value: %w", err)
				}
			case reflect.Invalid:
				return fmt.Errorf("invalid type in config: %v", e.field.Type)
			case reflect.Chan, reflect.Func, reflect.UnsafePointer:
				return fmt.Errorf("unsupported type in config: %v", e.field.Type)
			case reflect.Complex64, reflect.Complex128:
				return fmt.Errorf("complex types are not supported")
			case reflect.Pointer, reflect.Uintptr:
				return fmt.Errorf("pointer types are not supported")
			case reflect.Struct:
				// a struct itself can't be expressed in an environment variable, so we'll never get here
				return fmt.Errorf("unsupported struct type in config: %v", e.field.Type)
			default:
				return fmt.Errorf("unsupported type in config: %v", e.field.Type)
			}
		}
	}

	return nil
}

type env struct {
	name  string
	field reflect.StructField
	tag   tags.StructTag
}

func buildEnvs(typ reflect.Type, prefix, separator, arraySeparator string) ([]env, error) {
	envs := []env{}
	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Type.Kind() == reflect.Struct {
			newPrefix := prefix + field.Name + separator
			subenvs, err := buildEnvs(field.Type, newPrefix, separator, arraySeparator)
			if err != nil {
				return nil, err
			}
			envs = append(envs, subenvs...)
		} else if tag := field.Tag.Get("name"); tag != "" {
			tagInfo, err := tags.ExtractStructTags(field, arraySeparator)
			if err != nil {
				return nil, err
			}
			envs = append(envs, env{name: strings.ToUpper(prefix + tagInfo.Name), field: field, tag: tagInfo})
		}
	}
	return envs, nil
}
