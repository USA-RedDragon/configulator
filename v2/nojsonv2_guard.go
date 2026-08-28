//go:build !goexperiment.jsonv2

package configulator

// This file compiles ONLY when GOEXPERIMENT=nojsonv2 has disabled
// encoding/json/v2. Without this file, that build would fail with
// the stdlib's "build constraints exclude all Go files in
// .../encoding/json/v2", which isn't friendly.
var _ = configulator_v2_requires_encoding_json_v2
