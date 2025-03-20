package configulator

import (
	"context"
	"testing"

	"github.com/spf13/pflag"
)

type testConfig struct {
	Bool          bool          `config:"bool,default:true,description:bool"`
	Int           int           `config:"int,default:1,description:int"`
	Int8          int8          `config:"int8,default:2,description:int8"`
	Int16         int16         `config:"int16,default:3,description:int16"`
	Int32         int32         `config:"int32,default:4,description:int32"`
	Int64         int64         `config:"int64,default:5,description:int64"`
	Uint          uint          `config:"uint,default:6,description:uint"`
	Uint8         uint8         `config:"uint8,default:7,description:uint8"`
	Uint16        uint16        `config:"uint16,default:8,description:uint16"`
	Uint32        uint32        `config:"uint32,default:9,description:uint32"`
	Uint64        uint64        `config:"uint64,default:10,description:uint64"`
	Float32       float32       `config:"float32,default:11.0,description:float32"`
	Float64       float64       `config:"float64,default:12.0,description:float64"`
	String        string        `config:"string,default:15.0,description:string"`
	SubTestConfig subTestConfig `config:"subTestConfig,description:subTestConfig"`
	Unexported    int
	// Known issues:
	// defaults in arrays are not split appropriately
	// StringArray []string `config:"stringArray,default:16.0,14.0,description:array"`
	// arrays of structs are not yet implemented
	// SubTestConfigArray []subTestConfig `config:"subTestConfigArray,description:subTestConfigArray"`
}

func (c testConfig) Validate() error {
	return nil
}

type subTestConfig struct {
	Bool       bool    `config:"bool,default:true,description:bool"`
	Int        int     `config:"int,default:1,description:int"`
	Int8       int8    `config:"int8,default:2,description:int8"`
	Int16      int16   `config:"int16,default:3,description:int16"`
	Int32      int32   `config:"int32,default:4,description:int32"`
	Int64      int64   `config:"int64,default:5,description:int64"`
	Uint       uint    `config:"uint,default:6,description:uint"`
	Uint8      uint8   `config:"uint8,default:7,description:uint8"`
	Uint16     uint16  `config:"uint16,default:8,description:uint16"`
	Uint32     uint32  `config:"uint32,default:9,description:uint32"`
	Uint64     uint64  `config:"uint64,default:10,description:uint64"`
	Float32    float32 `config:"float32,default:11.0,description:float32"`
	Float64    float64 `config:"float64,default:12.0,description:float64"`
	String     string  `config:"string,default:15.0,description:string"`
	Unexported int
	// Known issues:
	// defaults in arrays are not split appropriately
	// StringArray []string `config:"stringArray,default:16.0,14.0,description:array"`
	// arrays of structs are not yet implemented
	// SubTestConfigArray []subTestConfig `config:"subTestConfigArray,description:subTestConfigArray"`
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
	cFromCtx, err := FromContext[testConfig](subCtx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cFromCtx == nil {
		t.Fatal("expected non-nil Configulator from context")
	}
	if cFromCtx != c {
		t.Fatal("expected Configulator from context to be the same as original")
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

func TestConfigulatorEnvironmentVariables(t *testing.T) {
	t.Setenv("TEST_BOOL", "false")
	t.Setenv("TEST_INT", "20")
	t.Setenv("TEST_INT8", "21")
	t.Setenv("TEST_INT16", "22")
	t.Setenv("TEST_INT32", "23")
	t.Setenv("TEST_INT64", "24")
	t.Setenv("TEST_UINT", "25")
	t.Setenv("TEST_UINT8", "26")
	t.Setenv("TEST_UINT16", "27")
	t.Setenv("TEST_UINT32", "28")
	t.Setenv("TEST_UINT64", "29")
	t.Setenv("TEST_FLOAT32", "30.0")
	t.Setenv("TEST_FLOAT64", "31.0")
	t.Setenv("TEST_STRING", "32.0")
	t.Setenv("TEST_SUBTESTCONFIG_BOOL", "false")
	t.Setenv("TEST_SUBTESTCONFIG_INT", "40")
	t.Setenv("TEST_SUBTESTCONFIG_INT8", "41")
	t.Setenv("TEST_SUBTESTCONFIG_INT16", "42")
	t.Setenv("TEST_SUBTESTCONFIG_INT32", "43")
	t.Setenv("TEST_SUBTESTCONFIG_INT64", "44")
	t.Setenv("TEST_SUBTESTCONFIG_UINT", "45")
	t.Setenv("TEST_SUBTESTCONFIG_UINT8", "46")
	t.Setenv("TEST_SUBTESTCONFIG_UINT16", "47")
	t.Setenv("TEST_SUBTESTCONFIG_UINT32", "48")
	t.Setenv("TEST_SUBTESTCONFIG_UINT64", "49")
	t.Setenv("TEST_SUBTESTCONFIG_FLOAT32", "50.0")
	t.Setenv("TEST_SUBTESTCONFIG_FLOAT64", "51.0")
	t.Setenv("TEST_SUBTESTCONFIG_STRING", "52.0")

	c := New[testConfig]()
	c.WithEnvironmentVariables(&EnvironmentVariableOptions{
		Prefix:    "TEST_",
		Separator: "_",
	})

	cfg, err := c.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Bool != false {
		t.Fatalf("expected Bool to be false, got %v", cfg.Bool)
	}

	if cfg.Int != 20 {
		t.Fatalf("expected Int to be 20, got %v", cfg.Int)
	}

	if cfg.Int8 != 21 {
		t.Fatalf("expected Int8 to be 21, got %v", cfg.Int8)
	}

	if cfg.Int16 != 22 {
		t.Fatalf("expected Int16 to be 22, got %v", cfg.Int16)
	}

	if cfg.Int32 != 23 {
		t.Fatalf("expected Int32 to be 23, got %v", cfg.Int32)
	}

	if cfg.Int64 != 24 {
		t.Fatalf("expected Int64 to be 24, got %v", cfg.Int64)
	}

	if cfg.Uint != 25 {
		t.Fatalf("expected Uint to be 25, got %v", cfg.Uint)
	}

	if cfg.Uint8 != 26 {
		t.Fatalf("expected Uint8 to be 26, got %v", cfg.Uint8)
	}

	if cfg.Uint16 != 27 {
		t.Fatalf("expected Uint16 to be 27, got %v", cfg.Uint16)
	}

	if cfg.Uint32 != 28 {
		t.Fatalf("expected Uint32 to be 28, got %v", cfg.Uint32)
	}

	if cfg.Uint64 != 29 {
		t.Fatalf("expected Uint64 to be 29, got %v", cfg.Uint64)
	}

	if cfg.Float32 != 30.0 {
		t.Fatalf("expected Float32 to be 30.0, got %v", cfg.Float32)
	}

	if cfg.Float64 != 31.0 {
		t.Fatalf("expected Float64 to be 31.0, got %v", cfg.Float64)
	}

	if cfg.String != "32.0" {
		t.Fatalf("expected String to be '32.0', got '%s'", cfg.String)
	}

	if cfg.SubTestConfig.Bool != false {
		t.Fatalf("expected SubTestConfig.Bool to be false, got %v", cfg.SubTestConfig.Bool)
	}

	if cfg.SubTestConfig.Int != 40 {
		t.Fatalf("expected SubTestConfig.Int to be 40, got %v", cfg.SubTestConfig.Int)
	}

	if cfg.SubTestConfig.Int8 != 41 {
		t.Fatalf("expected SubTestConfig.Int8 to be 41, got %v", cfg.SubTestConfig.Int8)
	}

	if cfg.SubTestConfig.Int16 != 42 {
		t.Fatalf("expected SubTestConfig.Int16 to be 42, got %v", cfg.SubTestConfig.Int16)
	}

	if cfg.SubTestConfig.Int32 != 43 {
		t.Fatalf("expected SubTestConfig.Int32 to be 43, got %v", cfg.SubTestConfig.Int32)
	}

	if cfg.SubTestConfig.Int64 != 44 {
		t.Fatalf("expected SubTestConfig.Int64 to be 44, got %v", cfg.SubTestConfig.Int64)
	}

	if cfg.SubTestConfig.Uint != 45 {
		t.Fatalf("expected SubTestConfig.Uint to be 45, got %v", cfg.SubTestConfig.Uint)
	}

	if cfg.SubTestConfig.Uint8 != 46 {
		t.Fatalf("expected SubTestConfig.Uint8 to be 46, got %v", cfg.SubTestConfig.Uint8)
	}

	if cfg.SubTestConfig.Uint16 != 47 {
		t.Fatalf("expected SubTestConfig.Uint16 to be 47, got %v", cfg.SubTestConfig.Uint16)
	}

	if cfg.SubTestConfig.Uint32 != 48 {
		t.Fatalf("expected SubTestConfig.Uint32 to be 48, got %v", cfg.SubTestConfig.Uint32)
	}

	if cfg.SubTestConfig.Uint64 != 49 {
		t.Fatalf("expected SubTestConfig.Uint64 to be 49, got %v", cfg.SubTestConfig.Uint64)
	}

	if cfg.SubTestConfig.Float32 != 50.0 {
		t.Fatalf("expected SubTestConfig.Float32 to be 50.0, got %v", cfg.SubTestConfig.Float32)
	}

	if cfg.SubTestConfig.Float64 != 51.0 {
		t.Fatalf("expected SubTestConfig.Float64 to be 51.0, got %v", cfg.SubTestConfig.Float64)
	}

	if cfg.SubTestConfig.String != "52.0" {
		t.Fatalf("expected SubTestConfig.String to be '52.0', got '%s'", cfg.SubTestConfig.String)
	}

}
