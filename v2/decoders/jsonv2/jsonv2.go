//go:build goexperiment.jsonv2

// Package jsonv2 has JSON decoders for FileOptions.Decoders, built on
// encoding/json/v2.
package jsonv2

import (
	jsonv2 "encoding/json/v2"

	configulator "github.com/USA-RedDragon/configulator/v2"
)

// Strict rejects unknown members.
func Strict(b []byte, v any) error {
	return configulator.StrictJSON(b, v)
}

// Lenient ignores unknown members.
func Lenient(b []byte, v any) error {
	return jsonv2.Unmarshal(b, v)
}

var (
	_ configulator.Unmarshal = Strict
	_ configulator.Unmarshal = Lenient
)
