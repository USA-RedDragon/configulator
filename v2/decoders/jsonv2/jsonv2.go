//go:build goexperiment.jsonv2

package jsonv2

import (
	jsonv2 "encoding/json/v2"

	configulator "github.com/USA-RedDragon/configulator/v2"
)

// Strict rejects unknown members
var Strict configulator.Unmarshal = configulator.StrictJSON

// Lenient ignores unknown members
var Lenient configulator.Unmarshal = func(b []byte, v any) error {
	return jsonv2.Unmarshal(b, v)
}
