//go:build goexperiment.jsonv2

package conformance

import (
	"testing"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/configulator/v2/decoders/jsonv2"
)

// TestUnknownKeysFollowDecoder checks that the generated JSON code skips
// unknown keys at every level under jsonv2.Lenient and rejects them under
// the strict decoders.
func TestUnknownKeysFollowDecoder(t *testing.T) {
	t.Parallel()
	const data = `{"extra":{"a":[1,{"b":2}]},"http":{"port":9000,"extra":true},"db":{"pool":{"size":3,"extra":"x"}}}`
	load := func(u configulator.Unmarshal) (*Nested, error) {
		return configulator.New(NestedSchema()).
			WithFile(&configulator.FileOptions{Explicit: "c.json", Decoders: configulator.Decoders{".json": u}}).
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
