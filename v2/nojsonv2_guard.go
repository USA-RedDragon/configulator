//go:build !goexperiment.jsonv2

package configulator

// This file only builds with GOEXPERIMENT=nojsonv2. The undefined name
// below gives a clearer error than "build constraints exclude all Go files
// in .../encoding/json/v2".
var _ = configulator_v2_requires_encoding_json_v2
