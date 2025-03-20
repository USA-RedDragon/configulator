package configulator

import (
	"context"
	"testing"

	"github.com/spf13/pflag"
)

type testConfig struct {
	Bool    bool    `config:"bool,default:true,description:bool"`
	Int     int     `config:"int,default:1,description:int"`
	Int8    int8    `config:"int8,default:2,description:int8"`
	Int16   int16   `config:"int16,default:3,description:int16"`
	Int32   int32   `config:"int32,default:4,description:int32"`
	Int64   int64   `config:"int64,default:5,description:int64"`
	Uint    uint    `config:"uint,default:6,description:uint"`
	Uint8   uint8   `config:"uint8,default:7,description:uint8"`
	Uint16  uint16  `config:"uint16,default:8,description:uint16"`
	Uint32  uint32  `config:"uint32,default:9,description:uint32"`
	Uint64  uint64  `config:"uint64,default:10,description:uint64"`
	Float32 float32 `config:"float32,default:11.0,description:float32"`
	Float64 float64 `config:"float64,default:12.0,description:float64"`
	String  string  `config:"string,default:15.0,description:string"`
	// StringArray []string `config:"stringArray,default:16.0,14.0,description:array"`
	Unexported int
}

func (c testConfig) Validate() error {
	return nil
}

func TestConfigulatorOptions(t *testing.T) {
	t.Parallel()

	pflags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	c := New[testConfig]()
	if c == nil {
		t.Fatal("expected non-nil Configulator")
	}
	c.WithPFlags(pflags, &PFlagOptions{
		Separator: "-",
	})
	if c.pflagOptions == nil {
		t.Fatal("expected non-nil PFlagOptions")
	}
	if c.pflagOptions.Separator != "-" {
		t.Fatalf("expected separator '-', got '%s'", c.pflagOptions.Separator)
	}
	if c.flags != pflags {
		t.Fatal("expected PFlags to be set")
	}

	c.WithArraySeparator(";")
	if c.arraySeparator != ";" {
		t.Fatalf("expected array separator ';', got '%s'", c.arraySeparator)
	}

	c.WithEnvironmentVariables(&EnvironmentVariableOptions{
		Prefix:    "TEST_",
		Separator: "_",
	})
	if c.envOptions == nil {
		t.Fatal("expected non-nil EnvironmentVariableOptions")
	}
	if c.envOptions.Prefix != "TEST_" {
		t.Fatalf("expected prefix 'TEST_', got '%s'", c.envOptions.Prefix)
	}
	if c.envOptions.Separator != "_" {
		t.Fatalf("expected separator '_', got '%s'", c.envOptions.Separator)
	}

	c.WithFile(&FileOptions{
		Paths: []string{"test.yaml"},
	})
	if c.fileOptions == nil {
		t.Fatal("expected non-nil FileOptions")
	}
	if len(c.fileOptions.Paths) != 1 {
		t.Fatal("expected one file path")
	}
	if c.fileOptions.Paths[0] != "test.yaml" {
		t.Fatalf("expected file path 'test.yaml', got '%s'", c.fileOptions.Paths[0])
	}

	ctx := context.TODO()
	subCtx := c.WithContext(ctx)
	if subCtx == nil {
		t.Fatal("expected non-nil context")
	}
	if subCtx.Value(configKey) != c {
		t.Fatal("expected context to contain Configulator")
	}
}

func TestConfigulatorDefaults(t *testing.T) {
	t.Parallel()

	c := New[testConfig]()
	if c == nil {
		t.Fatal("expected non-nil Configulator")
	}

	cfg, err := c.Default()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Bool != true {
		t.Fatalf("expected default Bool to be true, got %v", cfg.Bool)
	}

	if cfg.Int != 1 {
		t.Fatalf("expected default Int to be 1, got %v", cfg.Int)
	}

	if cfg.Int8 != 2 {
		t.Fatalf("expected default Int8 to be 2, got %v", cfg.Int8)
	}

	if cfg.Int16 != 3 {
		t.Fatalf("expected default Int16 to be 3, got %v", cfg.Int16)
	}

	if cfg.Int32 != 4 {
		t.Fatalf("expected default Int32 to be 4, got %v", cfg.Int32)
	}

	if cfg.Int64 != 5 {
		t.Fatalf("expected default Int64 to be 5, got %v", cfg.Int64)
	}

	if cfg.Uint != 6 {
		t.Fatalf("expected default Uint to be 6, got %v", cfg.Uint)
	}

	if cfg.Uint8 != 7 {
		t.Fatalf("expected default Uint8 to be 7, got %v", cfg.Uint8)
	}

	if cfg.Uint16 != 8 {
		t.Fatalf("expected default Uint16 to be 8, got %v", cfg.Uint16)
	}

	if cfg.Uint32 != 9 {
		t.Fatalf("expected default Uint32 to be 9, got %v", cfg.Uint32)
	}

	if cfg.Uint64 != 10 {
		t.Fatalf("expected default Uint64 to be 10, got %v", cfg.Uint64)
	}

	if cfg.Float32 != 11.0 {
		t.Fatalf("expected default Float32 to be 11.0, got %v", cfg.Float32)
	}

	if cfg.Float64 != 12.0 {
		t.Fatalf("expected default Float64 to be 12.0, got %v", cfg.Float64)
	}

	if cfg.String != "15.0" {
		t.Fatalf("expected default String to be '15.0', got '%s'", cfg.String)
	}

	// if len(cfg.StringArray) != 2 {
	// 	t.Fatalf("expected default StringArray to have 2 elements, got %d", len(cfg.StringArray))
	// }

	// if cfg.StringArray[0] != "16.0" {
	// 	t.Fatalf("expected default StringArray[0] to be '16.0', got '%s'", cfg.StringArray[0])
	// }

	// if cfg.StringArray[1] != "14.0" {
	// 	t.Fatalf("expected default StringArray[1] to be '14.0', got '%s'", cfg.StringArray[1])
	// }

	if c.cfg != nil {
		t.Fatal("expected cfg to be nil")
	}

}
