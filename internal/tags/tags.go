package tags

import (
	"fmt"
	"reflect"
	"strings"

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

	nameTag, err := ExtractNameFromTags(field.Tag)
	if err != nil {
		return ret, fmt.Errorf("field %s: %w", field.Name, err)
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

func ExtractNameFromTags(tag reflect.StructTag) (string, error) {
	// Check tags in order of precedence
	// 1. name
	// 2. json
	// 3. yaml
	nameTag := tag.Get("name")
	if nameTag == "" {
		nameTag = tag.Get("json")
		nameTag = strings.SplitN(nameTag, ",", 2)[0]
		if nameTag == "" {
			nameTag = tag.Get("yaml")
			nameTag = strings.SplitN(nameTag, ",", 2)[0]
			if nameTag == "" {
				return "", fmt.Errorf("no name, json, or yaml tag found")
			}
		}
	}
	return nameTag, nil
}
