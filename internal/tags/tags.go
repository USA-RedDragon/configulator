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

func ExtractStructTag(field reflect.StructField, tag, arraySeparator string) (StructTag, error) {
	ret := StructTag{}
	tags := strings.Split(tag, ",")
	if len(tags) == 0 {
		return ret, fmt.Errorf("invalid tag")
	}

	setDefault := false

	for _, tag := range tags {
		parts := strings.Split(tag, ":")
		if len(parts) > 2 {
			// recombine parts 1-end into a single string
			parts = append([]string{parts[0]}, strings.Join(parts[1:], ":"))
		}

		if len(parts) == 1 {
			ret.Name = parts[0]
		} else if len(parts) == 2 {
			if parts[0] == "name" {
				ret.Name = parts[1]
			}
			if parts[0] == "description" {
				ret.Description = parts[1]
			}
			if parts[0] == "default" {
				def, err := wrapper.WrapString(field.Type, parts[1], arraySeparator)
				if err != nil {
					return ret, err
				}
				ret.DefaultVal = def
				setDefault = true
			}
		}
	}

	if ret.Name == "" {
		return ret, fmt.Errorf("missing name")
	}

	if !setDefault {
		ret.DefaultVal = wrapper.WrappedValue{}
	}

	return ret, nil
}
