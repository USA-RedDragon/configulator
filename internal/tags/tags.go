package tags

import (
	"fmt"
	"reflect"

	"github.com/USA-RedDragon/configulator/internal/wrapper"
)

type StructTag struct {
	Name        string
	Description string
	DefaultVal  wrapper.WrappedValue
}

func ExtractStructTags(field reflect.StructField, arraySeparator string) (StructTag, error) {
	ret := StructTag{
		Description: field.Tag.Get("description"),
	}

	nameTag := field.Tag.Get("name")
	if nameTag == "" {
		return ret, fmt.Errorf("missing name tag")
	}
	ret.Name = nameTag

	defaultTag := field.Tag.Get("default")
	if defaultTag != "" {
		def, err := wrapper.WrapString(field.Type, defaultTag, arraySeparator)
		if err != nil {
			return ret, err
		}
		ret.DefaultVal = def
	}

	return ret, nil
}
