//go:build goexperiment.jsonv2

package conformance

import (
	"errors"
	"strings"
	"testing"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/configulator/v2/decoders/jsonv2"
)

// jsonFile is the name the tests here load their JSON as.
const jsonFile = "c.json"

// TestUnknownKeysFollowDecoder checks that the generated JSON code skips
// unknown keys at every level under jsonv2.Lenient and rejects them under
// the strict decoders.
func TestUnknownKeysFollowDecoder(t *testing.T) {
	t.Parallel()
	const data = `{"extra":{"a":[1,{"b":2}]},"http":{"port":9000,"extra":true},"db":{"pool":{"size":3,"extra":"x"}}}`
	load := func(u configulator.Unmarshal) (*Nested, error) {
		return configulator.New(NestedSchema()).
			WithFile(&configulator.FileOptions{Explicit: jsonFile, Decoders: configulator.Decoders{".json": u}}).
			WithReadFile(func(string) ([]byte, error) { return []byte(data), nil }).
			Load()
	}
	cfg, err := load(jsonv2.Lenient)
	if err != nil {
		t.Fatalf("Lenient: %v", err)
	}
	if cfg.HTTP.Port != 9000 || cfg.DB.Pool.Size != 3 {
		t.Errorf("Lenient: got port %d and pool size %d, want 9000 and 3", cfg.HTTP.Port, cfg.DB.Pool.Size)
	}
	for name, u := range map[string]configulator.Unmarshal{"StrictJSON": configulator.StrictJSON, "jsonv2.Strict": jsonv2.Strict} {
		if _, err := load(u); err == nil {
			t.Errorf("%s accepted unknown keys", name)
		}
	}
}

// loadJSON loads data as the explicit file c.json into the schema s with
// StrictJSON.
func loadJSON[C any](s *configulator.Schema[C], data string) error {
	_, err := configulator.New(s).
		WithFile(&configulator.FileOptions{Explicit: jsonFile}).
		WithReadFile(func(string) ([]byte, error) { return []byte(data), nil }).
		Load()
	return err
}

// TestJSONErrorPaths checks that errors from the generated JSON code name
// the full dotted path, the same one origins use, as typed errors.
func TestJSONErrorPaths(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, data, path, value string
		load                    func(string) error
	}{
		{"nested", `{"http":{"port":"x"}}`, "http.port", "x", func(d string) error { return loadJSON(NestedSchema(), d) }},
		{"deep", `{"db":{"pool":{"size":70000}}}`, "db.pool.size", "70000", func(d string) error { return loadJSON(NestedSchema(), d) }},
		{"list-element", `{"servers":[{"addr":"a"},{"weight":"w"}]}`, "servers[1].weight", "w", func(d string) error { return loadJSON(CollectionsSchema(), d) }},
		{"map-element", `{"pools":{"a.b":{"size":true}}}`, `pools."a.b".size`, "true", func(d string) error { return loadJSON(CollectionsSchema(), d) }},
		{"map-scalar", `{"labels":{"k":1}}`, "labels.k", "1", func(d string) error { return loadJSON(CollectionsSchema(), d) }},
		{"list-scalar", `{"tags":["a",2]}`, "tags[1]", "2", func(d string) error { return loadJSON(CollectionsSchema(), d) }},
		{"not-object", `{"http":[1]}`, "http", "[", func(d string) error { return loadJSON(NestedSchema(), d) }},
		{"secret", `{"token":"hunter2"}`, "token", "(redacted)", func(d string) error { return loadJSON(AttributesSchema(), d) }},
	} {
		err := tc.load(tc.data)
		var de *configulator.DecodeError
		var pe *configulator.ParseError
		if !errors.As(err, &de) || !errors.As(err, &pe) {
			t.Errorf("%s: got %v, want a DecodeError holding a ParseError", tc.name, err)
			continue
		}
		if pe.Path != tc.path || pe.Value != tc.value || pe.Source != jsonFile {
			t.Errorf("%s: got Path %q Value %q Source %q, want %q, %q and c.json", tc.name, pe.Path, pe.Value, pe.Source, tc.path, tc.value)
		}
		if strings.Contains(err.Error(), "hunter2") {
			t.Errorf("%s: error leaks the secret: %v", tc.name, err)
		}
	}
	for _, tc := range []struct{ data, path string }{
		{`{"db":{"pool":{"zz":1}}}`, "db.pool.zz"},
		{`{"x.y":1}`, `"x.y"`},
	} {
		var ue *configulator.UnknownKeyError
		if err := loadJSON(NestedSchema(), tc.data); !errors.As(err, &ue) || ue.Path != tc.path {
			t.Errorf("%s: got %v, want an UnknownKeyError for %s", tc.data, err, tc.path)
		}
	}
}
