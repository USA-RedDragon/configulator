package configulator

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestFileNumericRange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{"uint16 overflow", "uint16: 70000\n", "overflows uint16"},
		{"uint8 negative", "uint8: -1\n", "negative"},
		{"int8 overflow", "int8: 200\n", "overflows int8"},
		{"int16 fractional", "int16: 1.5\n", "not an integer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := t.TempDir() + "/config.yaml"
			if err := os.WriteFile(path, []byte(tc.yaml), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := New[testConfig]().WithFile(&FileOptions{Paths: []string{path}}).LoadWithoutValidation()
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}

	path := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(path, []byte("uint16: 65535\nint8: -128\nuint8: 255\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := New[testConfig]().WithFile(&FileOptions{Paths: []string{path}}).LoadWithoutValidation()
	if err != nil {
		t.Fatalf("in-range values should load: %v", err)
	}
	if cfg.Uint16 != 65535 || cfg.Int8 != -128 || cfg.Uint8 != 255 {
		t.Errorf("got uint16=%d int8=%d uint8=%d, want 65535 -128 255", cfg.Uint16, cfg.Int8, cfg.Uint8)
	}
}

func TestFileNumbers(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/config.yaml"
	doc := "int: 3\nint64: -4\nuint64: 5\nfloat32: 2\nfloat64: 1.5\n" +
		"intArray: [7, 8]\nint8Array: [-1, 2]\nuint16Array: [65535]\nuint64Array: [9]\n" +
		"float32Array: [1, 2.5]\nfloat64Array: [3]\n"
	if err := os.WriteFile(path, []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := New[testConfig]().WithFile(&FileOptions{Paths: []string{path}}).LoadWithoutValidation()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	got := fmt.Sprint(cfg.Int, cfg.Int64, cfg.Uint64, cfg.Float32, cfg.Float64,
		cfg.IntArray, cfg.Int8Array, cfg.Uint16Array, cfg.Uint64Array, cfg.Float32Array, cfg.Float64Array)
	want := "3 -4 5 2 1.5 [7 8] [-1 2] [65535] [9] [1 2.5] [3]"
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}

	if err := os.WriteFile(path, []byte("uint16Array: [70000]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := New[testConfig]().WithFile(&FileOptions{Paths: []string{path}}).LoadWithoutValidation(); err == nil || !strings.Contains(err.Error(), "overflows uint16") {
		t.Errorf("err = %v, want an overflow error", err)
	}
}
