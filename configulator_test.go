package configulator

import (
	"context"
	"testing"

	"github.com/spf13/pflag"
)

type testConfig struct {
	Bool           bool          `name:"bool" default:"true" description:"bool"`
	Int            int           `name:"int" default:"1" description:"int"`
	Int8           int8          `name:"int8" default:"2" description:"int8"`
	Int16          int16         `name:"int16" default:"3" description:"int16"`
	Int32          int32         `name:"int32" default:"4" description:"int32"`
	Int64          int64         `name:"int64" default:"5" description:"int64"`
	Uint           uint          `name:"uint" default:"6" description:"uint"`
	Uint8          uint8         `name:"uint8" default:"7" description:"uint8"`
	Uint16         uint16        `name:"uint16" default:"8" description:"uint16"`
	Uint32         uint32        `name:"uint32" default:"9" description:"uint32"`
	Uint64         uint64        `name:"uint64" default:"10" description:"uint64"`
	Float32        float32       `name:"float32" default:"11.0" description:"float32"`
	Float64        float64       `name:"float64" default:"12.0" description:"float64"`
	String         string        `name:"string" default:"15.0" description:"string"`
	Any            any           `name:"any" default:"15.0" description:"any"`
	SubTestConfig  subTestConfig `name:"subTestConfig" description:"subTestConfig"`
	Unexported     int
	StringArray    []string  `name:"stringArray" default:"16.0,14.0" description:"array"`
	BoolArray      []bool    `name:"boolArray" default:"true,false" description:"bool array"`
	IntArray       []int     `name:"intArray" default:"1,2,3" description:"int array"`
	Int8Array      []int8    `name:"int8Array" default:"1,2,3" description:"int8 array"`
	Int16Array     []int16   `name:"int16Array" default:"1,2,3" description:"int16 array"`
	Int32Array     []int32   `name:"int32Array" default:"1,2,3" description:"int32 array"`
	Int64Array     []int64   `name:"int64Array" default:"1,2,3" description:"int64 array"`
	UintArray      []uint    `name:"uintArray" default:"1,2,3" description:"uint array"`
	Uint8Array     []uint8   `name:"uint8Array" default:"1,2,3" description:"uint8 array"`
	Uint16Array    []uint16  `name:"uint16Array" default:"1,2,3" description:"uint16 array"`
	Uint32Array    []uint32  `name:"uint32Array" default:"1,2,3" description:"uint32 array"`
	Uint64Array    []uint64  `name:"uint64Array" default:"1,2,3" description:"uint64 array"`
	Float32Array   []float32 `name:"float32Array" default:"1.0,2.0,3.0" description:"float32 array"`
	Float64Array   []float64 `name:"float64Array" default:"1.0,2.0,3.0" description:"float64 array"`
	InterfaceArray []any     `name:"interfaceArray" default:"1.0,2.0" description:"interface array"`
	// arrays of structs are not yet implemented
	// SubTestConfigArray []subTestConfig `name:"subTestConfigArray" description:"subTestConfig array"`
}

func (c testConfig) Validate() error {
	return nil
}

type subTestConfig struct {
	Bool           bool    `name:"bool" default:"true" description:"bool"`
	Int            int     `name:"int" default:"1" description:"int"`
	Int8           int8    `name:"int8" default:"2" description:"int8"`
	Int16          int16   `name:"int16" default:"3" description:"int16"`
	Int32          int32   `name:"int32" default:"4" description:"int32"`
	Int64          int64   `name:"int64" default:"5" description:"int64"`
	Uint           uint    `name:"uint" default:"6" description:"uint"`
	Uint8          uint8   `name:"uint8" default:"7" description:"uint8"`
	Uint16         uint16  `name:"uint16" default:"8" description:"uint16"`
	Uint32         uint32  `name:"uint32" default:"9" description:"uint32"`
	Uint64         uint64  `name:"uint64" default:"10" description:"uint64"`
	Float32        float32 `name:"float32" default:"11.0" description:"float32"`
	Float64        float64 `name:"float64" default:"12.0" description:"float64"`
	String         string  `name:"string" default:"15.0" description:"string"`
	Any            any     `name:"any" default:"15.0" description:"any"`
	Unexported     int
	StringArray    []string  `name:"stringArray" default:"16.0,14.0" description:"array"`
	BoolArray      []bool    `name:"boolArray" default:"true,false" description:"bool array"`
	IntArray       []int     `name:"intArray" default:"1,2,3" description:"int array"`
	Int8Array      []int8    `name:"int8Array" default:"1,2,3" description:"int8 array"`
	Int16Array     []int16   `name:"int16Array" default:"1,2,3" description:"int16 array"`
	Int32Array     []int32   `name:"int32Array" default:"1,2,3" description:"int32 array"`
	Int64Array     []int64   `name:"int64Array" default:"1,2,3" description:"int64 array"`
	UintArray      []uint    `name:"uintArray" default:"1,2,3" description:"uint array"`
	Uint8Array     []uint8   `name:"uint8Array" default:"1,2,3" description:"uint8 array"`
	Uint16Array    []uint16  `name:"uint16Array" default:"1,2,3" description:"uint16 array"`
	Uint32Array    []uint32  `name:"uint32Array" default:"1,2,3" description:"uint32 array"`
	Uint64Array    []uint64  `name:"uint64Array" default:"1,2,3" description:"uint64 array"`
	Float32Array   []float32 `name:"float32Array" default:"1.0,2.0,3.0" description:"float32 array"`
	Float64Array   []float64 `name:"float64Array" default:"1.0,2.0,3.0" description:"float64 array"`
	InterfaceArray []any     `name:"interfaceArray" default:"1.0,2.0" description:"interface array"`
	// arrays of structs are not yet implemented
	// SubTestConfigArray []subTestConfig `name:"subTestConfigArray" description:"subTestConfig array"`
}

type nonDefaultsTestConfig struct {
	Bool           bool                     `name:"bool" description:"bool"`
	Int            int                      `name:"int" description:"int"`
	Int8           int8                     `name:"int8" description:"int8"`
	Int16          int16                    `name:"int16" description:"int16"`
	Int32          int32                    `name:"int32" description:"int32"`
	Int64          int64                    `name:"int64" description:"int64"`
	Uint           uint                     `name:"uint" description:"uint"`
	Uint8          uint8                    `name:"uint8" description:"uint8"`
	Uint16         uint16                   `name:"uint16" description:"uint16"`
	Uint32         uint32                   `name:"uint32" description:"uint32"`
	Uint64         uint64                   `name:"uint64" description:"uint64"`
	Float32        float32                  `name:"float32" description:"float32"`
	Float64        float64                  `name:"float64" description:"float64"`
	String         string                   `name:"string" description:"string"`
	Any            any                      `name:"any" description:"any"`
	SubTestConfig  subNonDefaultsTestConfig `name:"subTestConfig" description:"subTestConfig"`
	Unexported     int
	StringArray    []string  `name:"stringArray" description:"array"`
	BoolArray      []bool    `name:"boolArray" description:"bool array"`
	IntArray       []int     `name:"intArray" description:"int array"`
	Int8Array      []int8    `name:"int8Array" description:"int8 array"`
	Int16Array     []int16   `name:"int16Array" description:"int16 array"`
	Int32Array     []int32   `name:"int32Array" description:"int32 array"`
	Int64Array     []int64   `name:"int64Array" description:"int64 array"`
	UintArray      []uint    `name:"uintArray" description:"uint array"`
	Uint8Array     []uint8   `name:"uint8Array" description:"uint8 array"`
	Uint16Array    []uint16  `name:"uint16Array" description:"uint16 array"`
	Uint32Array    []uint32  `name:"uint32Array" description:"uint32 array"`
	Uint64Array    []uint64  `name:"uint64Array" description:"uint64 array"`
	Float32Array   []float32 `name:"float32Array" description:"float32 array"`
	Float64Array   []float64 `name:"float64Array" description:"float64 array"`
	InterfaceArray []any     `name:"interfaceArray" description:"interface array"`
	// arrays of structs are not yet implemented
	// SubTestConfigArray []subTestConfig `name:"subTestConfigArray" description:"subTestConfig array"`
}

func (c nonDefaultsTestConfig) Validate() error {
	return nil
}

type subNonDefaultsTestConfig struct {
	Bool           bool    `name:"bool" description:"bool"`
	Int            int     `name:"int" description:"int"`
	Int8           int8    `name:"int8" description:"int8"`
	Int16          int16   `name:"int16" description:"int16"`
	Int32          int32   `name:"int32" description:"int32"`
	Int64          int64   `name:"int64" description:"int64"`
	Uint           uint    `name:"uint" description:"uint"`
	Uint8          uint8   `name:"uint8" description:"uint8"`
	Uint16         uint16  `name:"uint16" description:"uint16"`
	Uint32         uint32  `name:"uint32" description:"uint32"`
	Uint64         uint64  `name:"uint64" description:"uint64"`
	Float32        float32 `name:"float32" description:"float32"`
	Float64        float64 `name:"float64" description:"float64"`
	String         string  `name:"string" description:"string"`
	Any            any     `name:"any" description:"any"`
	Unexported     int
	StringArray    []string  `name:"stringArray" description:"array"`
	BoolArray      []bool    `name:"boolArray" description:"bool array"`
	IntArray       []int     `name:"intArray" description:"int array"`
	Int8Array      []int8    `name:"int8Array" description:"int8 array"`
	Int16Array     []int16   `name:"int16Array" description:"int16 array"`
	Int32Array     []int32   `name:"int32Array" description:"int32 array"`
	Int64Array     []int64   `name:"int64Array" description:"int64 array"`
	UintArray      []uint    `name:"uintArray" description:"uint array"`
	Uint8Array     []uint8   `name:"uint8Array" description:"uint8 array"`
	Uint16Array    []uint16  `name:"uint16Array" description:"uint16 array"`
	Uint32Array    []uint32  `name:"uint32Array" description:"uint32 array"`
	Uint64Array    []uint64  `name:"uint64Array" description:"uint64 array"`
	Float32Array   []float32 `name:"float32Array" description:"float32 array"`
	Float64Array   []float64 `name:"float64Array" description:"float64 array"`
	InterfaceArray []any     `name:"interfaceArray" description:"interface array"`
	// arrays of structs are not yet implemented
	// SubTestConfigArray []subTestConfig `name:"subTestConfigArray" description:"subTestConfig array"`
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

//nolint:golint,gocyclo
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

	//nolint:golint,goconst
	if cfg.String != "15.0" {
		t.Fatalf("expected default String to be '15.0', got '%s'", cfg.String)
	}

	if cfg.Unexported != 0 {
		t.Fatalf("expected default Unexported to be 0, got %v", cfg.Unexported)
	}

	if cfg.Any != "15.0" {
		t.Fatalf("expected default Any to be '15.0', got '%s'", cfg.Any)
	}

	if len(cfg.InterfaceArray) != 2 {
		t.Fatalf("expected default InterfaceArray to have 2 elements, got %d", len(cfg.InterfaceArray))
	}

	if cfg.InterfaceArray[0] != "1.0" {
		t.Fatalf("expected default InterfaceArray[0] to be '1.0', got %v", cfg.InterfaceArray[0])
	}

	if cfg.InterfaceArray[1] != "2.0" {
		t.Fatalf("expected default InterfaceArray[1] to be '2.0', got %v", cfg.InterfaceArray[1])
	}

	if len(cfg.StringArray) != 2 {
		t.Fatalf("expected default StringArray to have 2 elements, got %d", len(cfg.StringArray))
	}

	if cfg.StringArray[0] != "16.0" {
		t.Fatalf("expected default StringArray[0] to be '16.0', got '%s'", cfg.StringArray[0])
	}

	if cfg.StringArray[1] != "14.0" {
		t.Fatalf("expected default StringArray[1] to be '14.0', got '%s'", cfg.StringArray[1])
	}

	if len(cfg.BoolArray) != 2 {
		t.Fatalf("expected default BoolArray to have 2 elements, got %d", len(cfg.BoolArray))
	}

	if cfg.BoolArray[0] != true {
		t.Fatalf("expected default BoolArray[0] to be true, got %v", cfg.BoolArray[0])
	}

	if cfg.BoolArray[1] != false {
		t.Fatalf("expected default BoolArray[1] to be false, got %v", cfg.BoolArray[1])
	}

	if len(cfg.IntArray) != 3 {
		t.Fatalf("expected default IntArray to have 3 elements, got %d", len(cfg.IntArray))
	}

	if cfg.IntArray[0] != 1 {
		t.Fatalf("expected default IntArray[0] to be 1, got %v", cfg.IntArray[0])
	}

	if cfg.IntArray[1] != 2 {
		t.Fatalf("expected default IntArray[1] to be 2, got %v", cfg.IntArray[1])
	}

	if cfg.IntArray[2] != 3 {
		t.Fatalf("expected default IntArray[2] to be 3, got %v", cfg.IntArray[2])
	}

	if len(cfg.Int8Array) != 3 {
		t.Fatalf("expected default Int8Array to have 3 elements, got %d", len(cfg.Int8Array))
	}

	if cfg.Int8Array[0] != 1 {
		t.Fatalf("expected default Int8Array[0] to be 1, got %v", cfg.Int8Array[0])
	}

	if cfg.Int8Array[1] != 2 {
		t.Fatalf("expected default Int8Array[1] to be 2, got %v", cfg.Int8Array[1])
	}

	if cfg.Int8Array[2] != 3 {
		t.Fatalf("expected default Int8Array[2] to be 3, got %v", cfg.Int8Array[2])
	}

	if len(cfg.Int16Array) != 3 {
		t.Fatalf("expected default Int16Array to have 3 elements, got %d", len(cfg.Int16Array))
	}

	if cfg.Int16Array[0] != 1 {
		t.Fatalf("expected default Int16Array[0] to be 1, got %v", cfg.Int16Array[0])
	}

	if cfg.Int16Array[1] != 2 {
		t.Fatalf("expected default Int16Array[1] to be 2, got %v", cfg.Int16Array[1])
	}

	if cfg.Int16Array[2] != 3 {
		t.Fatalf("expected default Int16Array[2] to be 3, got %v", cfg.Int16Array[2])
	}

	if len(cfg.Int32Array) != 3 {
		t.Fatalf("expected default Int32Array to have 3 elements, got %d", len(cfg.Int32Array))
	}

	if cfg.Int32Array[0] != 1 {
		t.Fatalf("expected default Int32Array[0] to be 1, got %v", cfg.Int32Array[0])
	}

	if cfg.Int32Array[1] != 2 {
		t.Fatalf("expected default Int32Array[1] to be 2, got %v", cfg.Int32Array[1])
	}

	if cfg.Int32Array[2] != 3 {
		t.Fatalf("expected default Int32Array[2] to be 3, got %v", cfg.Int32Array[2])
	}

	if len(cfg.Int64Array) != 3 {
		t.Fatalf("expected default Int64Array to have 3 elements, got %d", len(cfg.Int64Array))
	}

	if cfg.Int64Array[0] != 1 {
		t.Fatalf("expected default Int64Array[0] to be 1, got %v", cfg.Int64Array[0])
	}

	if cfg.Int64Array[1] != 2 {
		t.Fatalf("expected default Int64Array[1] to be 2, got %v", cfg.Int64Array[1])
	}

	if cfg.Int64Array[2] != 3 {
		t.Fatalf("expected default Int64Array[2] to be 3, got %v", cfg.Int64Array[2])
	}

	if len(cfg.UintArray) != 3 {
		t.Fatalf("expected default UintArray to have 3 elements, got %d", len(cfg.UintArray))
	}

	if cfg.UintArray[0] != 1 {
		t.Fatalf("expected default UintArray[0] to be 1, got %v", cfg.UintArray[0])
	}

	if cfg.UintArray[1] != 2 {
		t.Fatalf("expected default UintArray[1] to be 2, got %v", cfg.UintArray[1])
	}

	if cfg.UintArray[2] != 3 {
		t.Fatalf("expected default UintArray[2] to be 3, got %v", cfg.UintArray[2])
	}

	if len(cfg.Uint8Array) != 3 {
		t.Fatalf("expected default Uint8Array to have 3 elements, got %d", len(cfg.Uint8Array))
	}

	if cfg.Uint8Array[0] != 1 {
		t.Fatalf("expected default Uint8Array[0] to be 1, got %v", cfg.Uint8Array[0])
	}

	if cfg.Uint8Array[1] != 2 {
		t.Fatalf("expected default Uint8Array[1] to be 2, got %v", cfg.Uint8Array[1])
	}

	if cfg.Uint8Array[2] != 3 {
		t.Fatalf("expected default Uint8Array[2] to be 3, got %v", cfg.Uint8Array[2])
	}

	if len(cfg.Uint16Array) != 3 {
		t.Fatalf("expected default Uint16Array to have 3 elements, got %d", len(cfg.Uint16Array))
	}

	if cfg.Uint16Array[0] != 1 {
		t.Fatalf("expected default Uint16Array[0] to be 1, got %v", cfg.Uint16Array[0])
	}

	if cfg.Uint16Array[1] != 2 {
		t.Fatalf("expected default Uint16Array[1] to be 2, got %v", cfg.Uint16Array[1])
	}

	if cfg.Uint16Array[2] != 3 {
		t.Fatalf("expected default Uint16Array[2] to be 3, got %v", cfg.Uint16Array[2])
	}

	if len(cfg.Uint32Array) != 3 {
		t.Fatalf("expected default Uint32Array to have 3 elements, got %d", len(cfg.Uint32Array))
	}

	if cfg.Uint32Array[0] != 1 {
		t.Fatalf("expected default Uint32Array[0] to be 1, got %v", cfg.Uint32Array[0])
	}

	if cfg.Uint32Array[1] != 2 {
		t.Fatalf("expected default Uint32Array[1] to be 2, got %v", cfg.Uint32Array[1])
	}

	if cfg.Uint32Array[2] != 3 {
		t.Fatalf("expected default Uint32Array[2] to be 3, got %v", cfg.Uint32Array[2])
	}

	if len(cfg.Uint64Array) != 3 {
		t.Fatalf("expected default Uint64Array to have 3 elements, got %d", len(cfg.Uint64Array))
	}

	if cfg.Uint64Array[0] != 1 {
		t.Fatalf("expected default Uint64Array[0] to be 1, got %v", cfg.Uint64Array[0])
	}

	if cfg.Uint64Array[1] != 2 {
		t.Fatalf("expected default Uint64Array[1] to be 2, got %v", cfg.Uint64Array[1])
	}

	if cfg.Uint64Array[2] != 3 {
		t.Fatalf("expected default Uint64Array[2] to be 3, got %v", cfg.Uint64Array[2])
	}

	if len(cfg.Float32Array) != 3 {
		t.Fatalf("expected default Float32Array to have 3 elements, got %d", len(cfg.Float32Array))
	}

	if cfg.Float32Array[0] != 1.0 {
		t.Fatalf("expected default Float32Array[0] to be 1.0, got %v", cfg.Float32Array[0])
	}

	if cfg.Float32Array[1] != 2.0 {
		t.Fatalf("expected default Float32Array[1] to be 2.0, got %v", cfg.Float32Array[1])
	}

	if cfg.Float32Array[2] != 3.0 {
		t.Fatalf("expected default Float32Array[2] to be 3.0, got %v", cfg.Float32Array[2])
	}

	if len(cfg.Float64Array) != 3 {
		t.Fatalf("expected default Float64Array to have 3 elements, got %d", len(cfg.Float64Array))
	}

	if cfg.Float64Array[0] != 1.0 {
		t.Fatalf("expected default Float64Array[0] to be 1.0, got %v", cfg.Float64Array[0])
	}

	if cfg.Float64Array[1] != 2.0 {
		t.Fatalf("expected default Float64Array[1] to be 2.0, got %v", cfg.Float64Array[1])
	}

	if cfg.Float64Array[2] != 3.0 {
		t.Fatalf("expected default Float64Array[2] to be 3.0, got %v", cfg.Float64Array[2])
	}

	if cfg.SubTestConfig.Bool != true {
		t.Fatalf("expected default SubTestConfig.Bool to be true, got %v", cfg.SubTestConfig.Bool)
	}

	if cfg.SubTestConfig.Int != 1 {
		t.Fatalf("expected default SubTestConfig.Int to be 1, got %v", cfg.SubTestConfig.Int)
	}

	if cfg.SubTestConfig.Int8 != 2 {
		t.Fatalf("expected default SubTestConfig.Int8 to be 2, got %v", cfg.SubTestConfig.Int8)
	}

	if cfg.SubTestConfig.Int16 != 3 {
		t.Fatalf("expected default SubTestConfig.Int16 to be 3, got %v", cfg.SubTestConfig.Int16)
	}

	if cfg.SubTestConfig.Int32 != 4 {
		t.Fatalf("expected default SubTestConfig.Int32 to be 4, got %v", cfg.SubTestConfig.Int32)
	}

	if cfg.SubTestConfig.Int64 != 5 {
		t.Fatalf("expected default SubTestConfig.Int64 to be 5, got %v", cfg.SubTestConfig.Int64)
	}

	if cfg.SubTestConfig.Uint != 6 {
		t.Fatalf("expected default SubTestConfig.Uint to be 6, got %v", cfg.SubTestConfig.Uint)
	}

	if cfg.SubTestConfig.Uint8 != 7 {
		t.Fatalf("expected default SubTestConfig.Uint8 to be 7, got %v", cfg.SubTestConfig.Uint8)
	}

	if cfg.SubTestConfig.Uint16 != 8 {
		t.Fatalf("expected default SubTestConfig.Uint16 to be 8, got %v", cfg.SubTestConfig.Uint16)
	}

	if cfg.SubTestConfig.Uint32 != 9 {
		t.Fatalf("expected default SubTestConfig.Uint32 to be 9, got %v", cfg.SubTestConfig.Uint32)
	}

	if cfg.SubTestConfig.Uint64 != 10 {
		t.Fatalf("expected default SubTestConfig.Uint64 to be 10, got %v", cfg.SubTestConfig.Uint64)
	}

	if cfg.SubTestConfig.Float32 != 11.0 {
		t.Fatalf("expected default SubTestConfig.Float32 to be 11.0, got %v", cfg.SubTestConfig.Float32)
	}

	if cfg.SubTestConfig.Float64 != 12.0 {
		t.Fatalf("expected default SubTestConfig.Float64 to be 12.0, got %v", cfg.SubTestConfig.Float64)
	}

	if cfg.SubTestConfig.String != "15.0" {
		t.Fatalf("expected default SubTestConfig.String to be '15.0', got '%s'", cfg.SubTestConfig.String)
	}

	if len(cfg.SubTestConfig.StringArray) != 2 {
		t.Fatalf("expected default SubTestConfig.StringArray to have 2 elements, got %d", len(cfg.SubTestConfig.StringArray))
	}

	if cfg.SubTestConfig.StringArray[0] != "16.0" {
		t.Fatalf("expected default SubTestConfig.StringArray[0] to be '16.0', got '%s'", cfg.SubTestConfig.StringArray[0])
	}

	if cfg.SubTestConfig.StringArray[1] != "14.0" {
		t.Fatalf("expected default SubTestConfig.StringArray[1] to be '14.0', got '%s'", cfg.SubTestConfig.StringArray[1])
	}

	if len(cfg.SubTestConfig.BoolArray) != 2 {
		t.Fatalf("expected default SubTestConfig.BoolArray to have 2 elements, got %d", len(cfg.SubTestConfig.BoolArray))
	}

	if cfg.SubTestConfig.BoolArray[0] != true {
		t.Fatalf("expected default SubTestConfig.BoolArray[0] to be true, got %v", cfg.SubTestConfig.BoolArray[0])
	}

	if cfg.SubTestConfig.BoolArray[1] != false {
		t.Fatalf("expected default SubTestConfig.BoolArray[1] to be false, got %v", cfg.SubTestConfig.BoolArray[1])
	}

	if len(cfg.SubTestConfig.IntArray) != 3 {
		t.Fatalf("expected default SubTestConfig.IntArray to have 3 elements, got %d", len(cfg.SubTestConfig.IntArray))
	}

	if cfg.SubTestConfig.IntArray[0] != 1 {
		t.Fatalf("expected default SubTestConfig.IntArray[0] to be 1, got %v", cfg.SubTestConfig.IntArray[0])
	}

	if cfg.SubTestConfig.IntArray[1] != 2 {
		t.Fatalf("expected default SubTestConfig.IntArray[1] to be 2, got %v", cfg.SubTestConfig.IntArray[1])
	}

	if cfg.SubTestConfig.IntArray[2] != 3 {
		t.Fatalf("expected default SubTestConfig.IntArray[2] to be 3, got %v", cfg.SubTestConfig.IntArray[2])
	}

	if len(cfg.SubTestConfig.Int8Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Int8Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int8Array))
	}

	if cfg.SubTestConfig.Int8Array[0] != 1 {
		t.Fatalf("expected default SubTestConfig.Int8Array[0] to be 1, got %v", cfg.SubTestConfig.Int8Array[0])
	}

	if cfg.SubTestConfig.Int8Array[1] != 2 {
		t.Fatalf("expected default SubTestConfig.Int8Array[1] to be 2, got %v", cfg.SubTestConfig.Int8Array[1])
	}

	if cfg.SubTestConfig.Int8Array[2] != 3 {
		t.Fatalf("expected default SubTestConfig.Int8Array[2] to be 3, got %v", cfg.SubTestConfig.Int8Array[2])
	}

	if len(cfg.SubTestConfig.Int16Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Int16Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int16Array))
	}

	if cfg.SubTestConfig.Int16Array[0] != 1 {
		t.Fatalf("expected default SubTestConfig.Int16Array[0] to be 1, got %v", cfg.SubTestConfig.Int16Array[0])
	}

	if cfg.SubTestConfig.Int16Array[1] != 2 {
		t.Fatalf("expected default SubTestConfig.Int16Array[1] to be 2, got %v", cfg.SubTestConfig.Int16Array[1])
	}

	if cfg.SubTestConfig.Int16Array[2] != 3 {
		t.Fatalf("expected default SubTestConfig.Int16Array[2] to be 3, got %v", cfg.SubTestConfig.Int16Array[2])
	}

	if len(cfg.SubTestConfig.Int32Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Int32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int32Array))
	}

	if cfg.SubTestConfig.Int32Array[0] != 1 {
		t.Fatalf("expected default SubTestConfig.Int32Array[0] to be 1, got %v", cfg.SubTestConfig.Int32Array[0])
	}

	if cfg.SubTestConfig.Int32Array[1] != 2 {
		t.Fatalf("expected default SubTestConfig.Int32Array[1] to be 2, got %v", cfg.SubTestConfig.Int32Array[1])
	}

	if cfg.SubTestConfig.Int32Array[2] != 3 {
		t.Fatalf("expected default SubTestConfig.Int32Array[2] to be 3, got %v", cfg.SubTestConfig.Int32Array[2])
	}

	if len(cfg.SubTestConfig.Int64Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Int64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int64Array))
	}

	if cfg.SubTestConfig.Int64Array[0] != 1 {
		t.Fatalf("expected default SubTestConfig.Int64Array[0] to be 1, got %v", cfg.SubTestConfig.Int64Array[0])
	}

	if cfg.SubTestConfig.Int64Array[1] != 2 {
		t.Fatalf("expected default SubTestConfig.Int64Array[1] to be 2, got %v", cfg.SubTestConfig.Int64Array[1])
	}

	if cfg.SubTestConfig.Int64Array[2] != 3 {
		t.Fatalf("expected default SubTestConfig.Int64Array[2] to be 3, got %v", cfg.SubTestConfig.Int64Array[2])
	}

	if len(cfg.SubTestConfig.UintArray) != 3 {
		t.Fatalf("expected default SubTestConfig.UintArray to have 3 elements, got %d", len(cfg.SubTestConfig.UintArray))
	}

	if cfg.SubTestConfig.UintArray[0] != 1 {
		t.Fatalf("expected default SubTestConfig.UintArray[0] to be 1, got %v", cfg.SubTestConfig.UintArray[0])
	}

	if cfg.SubTestConfig.UintArray[1] != 2 {
		t.Fatalf("expected default SubTestConfig.UintArray[1] to be 2, got %v", cfg.SubTestConfig.UintArray[1])
	}

	if cfg.SubTestConfig.UintArray[2] != 3 {
		t.Fatalf("expected default SubTestConfig.UintArray[2] to be 3, got %v", cfg.SubTestConfig.UintArray[2])
	}

	if cfg.SubTestConfig.Uint8Array[0] != 1 {
		t.Fatalf("expected default SubTestConfig.Uint8Array[0] to be 1, got %v", cfg.SubTestConfig.Uint8Array[0])
	}

	if cfg.SubTestConfig.Uint8Array[1] != 2 {
		t.Fatalf("expected default SubTestConfig.Uint8Array[1] to be 2, got %v", cfg.SubTestConfig.Uint8Array[1])
	}

	if cfg.SubTestConfig.Uint8Array[2] != 3 {
		t.Fatalf("expected default SubTestConfig.Uint8Array[2] to be 3, got %v", cfg.SubTestConfig.Uint8Array[2])
	}

	if len(cfg.SubTestConfig.Uint16Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Uint16Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint16Array))
	}

	if cfg.SubTestConfig.Uint16Array[0] != 1 {
		t.Fatalf("expected default SubTestConfig.Uint16Array[0] to be 1, got %v", cfg.SubTestConfig.Uint16Array[0])
	}

	if cfg.SubTestConfig.Uint16Array[1] != 2 {
		t.Fatalf("expected default SubTestConfig.Uint16Array[1] to be 2, got %v", cfg.SubTestConfig.Uint16Array[1])
	}

	if cfg.SubTestConfig.Uint16Array[2] != 3 {
		t.Fatalf("expected default SubTestConfig.Uint16Array[2] to be 3, got %v", cfg.SubTestConfig.Uint16Array[2])
	}

	if len(cfg.SubTestConfig.Uint32Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Uint32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint32Array))
	}

	if cfg.SubTestConfig.Uint32Array[0] != 1 {
		t.Fatalf("expected default SubTestConfig.Uint32Array[0] to be 1, got %v", cfg.SubTestConfig.Uint32Array[0])
	}

	if cfg.SubTestConfig.Uint32Array[1] != 2 {
		t.Fatalf("expected default SubTestConfig.Uint32Array[1] to be 2, got %v", cfg.SubTestConfig.Uint32Array[1])
	}

	if cfg.SubTestConfig.Uint32Array[2] != 3 {
		t.Fatalf("expected default SubTestConfig.Uint32Array[2] to be 3, got %v", cfg.SubTestConfig.Uint32Array[2])
	}

	if len(cfg.SubTestConfig.Uint64Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Uint64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint64Array))
	}

	if cfg.SubTestConfig.Uint64Array[0] != 1 {
		t.Fatalf("expected default SubTestConfig.Uint64Array[0] to be 1, got %v", cfg.SubTestConfig.Uint64Array[0])
	}

	if cfg.SubTestConfig.Uint64Array[1] != 2 {
		t.Fatalf("expected default SubTestConfig.Uint64Array[1] to be 2, got %v", cfg.SubTestConfig.Uint64Array[1])
	}

	if cfg.SubTestConfig.Uint64Array[2] != 3 {
		t.Fatalf("expected default SubTestConfig.Uint64Array[2] to be 3, got %v", cfg.SubTestConfig.Uint64Array[2])
	}

	if len(cfg.SubTestConfig.Float32Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Float32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Float32Array))
	}

	if cfg.SubTestConfig.Float32Array[0] != 1.0 {
		t.Fatalf("expected default SubTestConfig.Float32Array[0] to be 1.0, got %v", cfg.SubTestConfig.Float32Array[0])
	}

	if cfg.SubTestConfig.Float32Array[1] != 2.0 {
		t.Fatalf("expected default SubTestConfig.Float32Array[1] to be 2.0, got %v", cfg.SubTestConfig.Float32Array[1])
	}

	if cfg.SubTestConfig.Float32Array[2] != 3.0 {
		t.Fatalf("expected default SubTestConfig.Float32Array[2] to be 3.0, got %v", cfg.SubTestConfig.Float32Array[2])
	}

	if len(cfg.SubTestConfig.Float64Array) != 3 {
		t.Fatalf("expected default SubTestConfig.Float64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Float64Array))
	}

	if cfg.SubTestConfig.Float64Array[0] != 1.0 {
		t.Fatalf("expected default SubTestConfig.Float64Array[0] to be 1.0, got %v", cfg.SubTestConfig.Float64Array[0])
	}

	if cfg.SubTestConfig.Float64Array[1] != 2.0 {
		t.Fatalf("expected default SubTestConfig.Float64Array[1] to be 2.0, got %v", cfg.SubTestConfig.Float64Array[1])
	}

	if cfg.SubTestConfig.Float64Array[2] != 3.0 {
		t.Fatalf("expected default SubTestConfig.Float64Array[2] to be 3.0, got %v", cfg.SubTestConfig.Float64Array[2])
	}

	if cfg.SubTestConfig.Unexported != 0 {
		t.Fatalf("expected default SubTestConfig.Unexported to be 0, got %v", cfg.SubTestConfig.Unexported)
	}

	if cfg.SubTestConfig.Any != "15.0" {
		t.Fatalf("expected default SubTestConfig.Any to be '15.0', got '%s'", cfg.SubTestConfig.Any)
	}

	if len(cfg.SubTestConfig.InterfaceArray) != 2 {
		t.Fatalf("expected default SubTestConfig.InterfaceArray to have 2 elements, got %d", len(cfg.SubTestConfig.InterfaceArray))
	}

	if cfg.SubTestConfig.InterfaceArray[0] != "1.0" {
		t.Fatalf("expected default SubTestConfig.InterfaceArray[0] to be '1.0', got %v", cfg.SubTestConfig.InterfaceArray[0])
	}

	if cfg.SubTestConfig.InterfaceArray[1] != "2.0" {
		t.Fatalf("expected default SubTestConfig.InterfaceArray[1] to be '2.0', got %v", cfg.SubTestConfig.InterfaceArray[1])
	}

	if c.cfg != nil {
		t.Fatal("expected cfg to be nil")
	}
}

//nolint:golint,gocyclo
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
	t.Setenv("TEST_ANY", "33.0")
	t.Setenv("TEST_STRINGARRAY", "53.0,54.0")
	t.Setenv("TEST_BOOLARRAY", "false,false,true")
	t.Setenv("TEST_INTARRAY", "33,34,35")
	t.Setenv("TEST_INT8ARRAY", "36,37,38")
	t.Setenv("TEST_INT16ARRAY", "39,40,41")
	t.Setenv("TEST_INT32ARRAY", "42,43,44")
	t.Setenv("TEST_INT64ARRAY", "45,46,47")
	t.Setenv("TEST_UINTARRAY", "48,49,50")
	t.Setenv("TEST_UINT8ARRAY", "51,52,53")
	t.Setenv("TEST_UINT16ARRAY", "54,55,56")
	t.Setenv("TEST_UINT32ARRAY", "57,58,59")
	t.Setenv("TEST_UINT64ARRAY", "60,61,62")
	t.Setenv("TEST_FLOAT32ARRAY", "63.0,64.0,65.0")
	t.Setenv("TEST_FLOAT64ARRAY", "66.0,67.0,68.0")
	t.Setenv("TEST_INTERFACEARRAY", "5.0,8.0")

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
	t.Setenv("TEST_SUBTESTCONFIG_ANY", "53.0")
	t.Setenv("TEST_SUBTESTCONFIG_STRINGARRAY", "68,69")
	t.Setenv("TEST_SUBTESTCONFIG_BOOLARRAY", "false,false,true")
	t.Setenv("TEST_SUBTESTCONFIG_INTARRAY", "53,54,55")
	t.Setenv("TEST_SUBTESTCONFIG_INT8ARRAY", "56,57,58")
	t.Setenv("TEST_SUBTESTCONFIG_INT16ARRAY", "59,60,61")
	t.Setenv("TEST_SUBTESTCONFIG_INT32ARRAY", "62,63,64")
	t.Setenv("TEST_SUBTESTCONFIG_INT64ARRAY", "65,66,67")
	t.Setenv("TEST_SUBTESTCONFIG_UINTARRAY", "68,69,70")
	t.Setenv("TEST_SUBTESTCONFIG_UINT8ARRAY", "71,72,73")
	t.Setenv("TEST_SUBTESTCONFIG_UINT16ARRAY", "74,75,76")
	t.Setenv("TEST_SUBTESTCONFIG_UINT32ARRAY", "77,78,79")
	t.Setenv("TEST_SUBTESTCONFIG_UINT64ARRAY", "80,81,82")
	t.Setenv("TEST_SUBTESTCONFIG_FLOAT32ARRAY", "83.0,84.0,85.0")
	t.Setenv("TEST_SUBTESTCONFIG_FLOAT64ARRAY", "86.0,87.0,88.0")
	t.Setenv("TEST_SUBTESTCONFIG_INTERFACEARRAY", "5.0,8.0")

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

	if cfg.Any != "33.0" {
		t.Fatalf("expected Any to be '33.0', got '%s'", cfg.Any)
	}

	if len(cfg.StringArray) != 2 {
		t.Fatalf("expected StringArray to have 2 elements, got %d", len(cfg.StringArray))
	}

	//nolint:goconst
	if cfg.StringArray[0] != "53.0" {
		t.Fatalf("expected StringArray[0] to be '53.0', got '%s'", cfg.StringArray[0])
	}

	if cfg.StringArray[1] != "54.0" {
		t.Fatalf("expected StringArray[1] to be '54.0', got '%s'", cfg.StringArray[1])
	}

	if len(cfg.BoolArray) != 3 {
		t.Fatalf("expected BoolArray to have 3 elements, got %d", len(cfg.BoolArray))
	}

	if cfg.BoolArray[0] != false {
		t.Fatalf("expected BoolArray[0] to be false, got %v", cfg.BoolArray[0])
	}

	if cfg.BoolArray[1] != false {
		t.Fatalf("expected BoolArray[1] to be false, got %v", cfg.BoolArray[1])
	}

	if cfg.BoolArray[2] != true {
		t.Fatalf("expected BoolArray[2] to be true, got %v", cfg.BoolArray[2])
	}

	if len(cfg.IntArray) != 3 {
		t.Fatalf("expected IntArray to have 3 elements, got %d", len(cfg.IntArray))
	}

	if cfg.IntArray[0] != 33 {
		t.Fatalf("expected IntArray[0] to be 33, got %v", cfg.IntArray[0])
	}

	if cfg.IntArray[1] != 34 {
		t.Fatalf("expected IntArray[1] to be 34, got %v", cfg.IntArray[1])
	}

	if cfg.IntArray[2] != 35 {
		t.Fatalf("expected IntArray[2] to be 35, got %v", cfg.IntArray[2])
	}

	if len(cfg.Int8Array) != 3 {
		t.Fatalf("expected Int8Array to have 3 elements, got %d", len(cfg.Int8Array))
	}

	if cfg.Int8Array[0] != 36 {
		t.Fatalf("expected Int8Array[0] to be 36, got %v", cfg.Int8Array[0])
	}

	if cfg.Int8Array[1] != 37 {
		t.Fatalf("expected Int8Array[1] to be 37, got %v", cfg.Int8Array[1])
	}

	if cfg.Int8Array[2] != 38 {
		t.Fatalf("expected Int8Array[2] to be 38, got %v", cfg.Int8Array[2])
	}

	if len(cfg.Int16Array) != 3 {
		t.Fatalf("expected Int16Array to have 3 elements, got %d", len(cfg.Int16Array))
	}

	if cfg.Int16Array[0] != 39 {
		t.Fatalf("expected Int16Array[0] to be 39, got %v", cfg.Int16Array[0])
	}

	if cfg.Int16Array[1] != 40 {
		t.Fatalf("expected Int16Array[1] to be 40, got %v", cfg.Int16Array[1])
	}

	if cfg.Int16Array[2] != 41 {
		t.Fatalf("expected Int16Array[2] to be 41, got %v", cfg.Int16Array[2])
	}

	if len(cfg.Int32Array) != 3 {
		t.Fatalf("expected Int32Array to have 3 elements, got %d", len(cfg.Int32Array))
	}

	if cfg.Int32Array[0] != 42 {
		t.Fatalf("expected Int32Array[0] to be 42, got %v", cfg.Int32Array[0])
	}

	if cfg.Int32Array[1] != 43 {
		t.Fatalf("expected Int32Array[1] to be 43, got %v", cfg.Int32Array[1])
	}

	if cfg.Int32Array[2] != 44 {
		t.Fatalf("expected Int32Array[2] to be 44, got %v", cfg.Int32Array[2])
	}

	if len(cfg.Int64Array) != 3 {
		t.Fatalf("expected Int64Array to have 3 elements, got %d", len(cfg.Int64Array))
	}

	if cfg.Int64Array[0] != 45 {
		t.Fatalf("expected Int64Array[0] to be 45, got %v", cfg.Int64Array[0])
	}

	if cfg.Int64Array[1] != 46 {
		t.Fatalf("expected Int64Array[1] to be 46, got %v", cfg.Int64Array[1])
	}

	if cfg.Int64Array[2] != 47 {
		t.Fatalf("expected Int64Array[2] to be 47, got %v", cfg.Int64Array[2])
	}

	if len(cfg.UintArray) != 3 {
		t.Fatalf("expected UintArray to have 3 elements, got %d", len(cfg.UintArray))
	}

	if cfg.UintArray[0] != 48 {
		t.Fatalf("expected UintArray[0] to be 48, got %v", cfg.UintArray[0])
	}

	if cfg.UintArray[1] != 49 {
		t.Fatalf("expected UintArray[1] to be 49, got %v", cfg.UintArray[1])
	}

	if cfg.UintArray[2] != 50 {
		t.Fatalf("expected UintArray[2] to be 50, got %v", cfg.UintArray[2])
	}

	if len(cfg.Uint8Array) != 3 {
		t.Fatalf("expected Uint8Array to have 3 elements, got %d", len(cfg.Uint8Array))
	}

	if cfg.Uint8Array[0] != 51 {
		t.Fatalf("expected Uint8Array[0] to be 51, got %v", cfg.Uint8Array[0])
	}

	if cfg.Uint8Array[1] != 52 {
		t.Fatalf("expected Uint8Array[1] to be 52, got %v", cfg.Uint8Array[1])
	}

	if cfg.Uint8Array[2] != 53 {
		t.Fatalf("expected Uint8Array[2] to be 53, got %v", cfg.Uint8Array[2])
	}

	if len(cfg.Uint16Array) != 3 {
		t.Fatalf("expected Uint16Array to have 3 elements, got %d", len(cfg.Uint16Array))
	}

	if cfg.Uint16Array[0] != 54 {
		t.Fatalf("expected Uint16Array[0] to be 54, got %v", cfg.Uint16Array[0])
	}

	if cfg.Uint16Array[1] != 55 {
		t.Fatalf("expected Uint16Array[1] to be 55, got %v", cfg.Uint16Array[1])
	}

	if cfg.Uint16Array[2] != 56 {
		t.Fatalf("expected Uint16Array[2] to be 56, got %v", cfg.Uint16Array[2])
	}

	if len(cfg.Uint32Array) != 3 {
		t.Fatalf("expected Uint32Array to have 3 elements, got %d", len(cfg.Uint32Array))
	}

	if cfg.Uint32Array[0] != 57 {
		t.Fatalf("expected Uint32Array[0] to be 57, got %v", cfg.Uint32Array[0])
	}

	if cfg.Uint32Array[1] != 58 {
		t.Fatalf("expected Uint32Array[1] to be 58, got %v", cfg.Uint32Array[1])
	}

	if cfg.Uint32Array[2] != 59 {
		t.Fatalf("expected Uint32Array[2] to be 59, got %v", cfg.Uint32Array[2])
	}

	if len(cfg.Uint64Array) != 3 {
		t.Fatalf("expected Uint64Array to have 3 elements, got %d", len(cfg.Uint64Array))
	}

	if cfg.Uint64Array[0] != 60 {
		t.Fatalf("expected Uint64Array[0] to be 60, got %v", cfg.Uint64Array[0])
	}

	if cfg.Uint64Array[1] != 61 {
		t.Fatalf("expected Uint64Array[1] to be 61, got %v", cfg.Uint64Array[1])
	}

	if cfg.Uint64Array[2] != 62 {
		t.Fatalf("expected Uint64Array[2] to be 62, got %v", cfg.Uint64Array[2])
	}

	if len(cfg.Float32Array) != 3 {
		t.Fatalf("expected Float32Array to have 3 elements, got %d", len(cfg.Float32Array))
	}

	if cfg.Float32Array[0] != 63.0 {
		t.Fatalf("expected Float32Array[0] to be 63.0, got %v", cfg.Float32Array[0])
	}

	if cfg.Float32Array[1] != 64.0 {
		t.Fatalf("expected Float32Array[1] to be 64.0, got %v", cfg.Float32Array[1])
	}

	if cfg.Float32Array[2] != 65.0 {
		t.Fatalf("expected Float32Array[2] to be 65.0, got %v", cfg.Float32Array[2])
	}

	if len(cfg.Float64Array) != 3 {
		t.Fatalf("expected Float64Array to have 3 elements, got %d", len(cfg.Float64Array))
	}

	if cfg.Float64Array[0] != 66.0 {
		t.Fatalf("expected Float64Array[0] to be 66.0, got %v", cfg.Float64Array[0])
	}

	if cfg.Float64Array[1] != 67.0 {
		t.Fatalf("expected Float64Array[1] to be 67.0, got %v", cfg.Float64Array[1])
	}

	if cfg.Float64Array[2] != 68.0 {
		t.Fatalf("expected Float64Array[2] to be 68.0, got %v", cfg.Float64Array[2])
	}

	if len(cfg.InterfaceArray) != 2 {
		t.Fatalf("expected InterfaceArray to have 2 elements, got %d", len(cfg.InterfaceArray))
	}

	if cfg.InterfaceArray[0] != "5.0" {
		t.Fatalf("expected InterfaceArray[0] to be '5.0', got %v", cfg.InterfaceArray[0])
	}

	if cfg.InterfaceArray[1] != "8.0" {
		t.Fatalf("expected InterfaceArray[1] to be '8.0', got %v", cfg.InterfaceArray[1])
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

	if cfg.SubTestConfig.Any != "53.0" {
		t.Fatalf("expected SubTestConfig.Any to be '53.0', got '%s'", cfg.SubTestConfig.Any)
	}

	if len(cfg.SubTestConfig.StringArray) != 2 {
		t.Fatalf("expected SubTestConfig.StringArray to have 2 elements, got %d", len(cfg.SubTestConfig.StringArray))
	}

	if cfg.SubTestConfig.StringArray[0] != "68" {
		t.Fatalf("expected SubTestConfig.StringArray[0] to be '68', got '%s'", cfg.SubTestConfig.StringArray[0])
	}

	if cfg.SubTestConfig.StringArray[1] != "69" {
		t.Fatalf("expected SubTestConfig.StringArray[1] to be '69', got '%s'", cfg.SubTestConfig.StringArray[1])
	}

	if len(cfg.SubTestConfig.BoolArray) != 3 {
		t.Fatalf("expected SubTestConfig.BoolArray to have 3 elements, got %d", len(cfg.SubTestConfig.BoolArray))
	}

	if len(cfg.SubTestConfig.IntArray) != 3 {
		t.Fatalf("expected SubTestConfig.IntArray to have 3 elements, got %d", len(cfg.SubTestConfig.IntArray))
	}

	if cfg.SubTestConfig.IntArray[0] != 53 {
		t.Fatalf("expected SubTestConfig.IntArray[0] to be 53, got %v", cfg.SubTestConfig.IntArray[0])
	}

	if cfg.SubTestConfig.IntArray[1] != 54 {
		t.Fatalf("expected SubTestConfig.IntArray[1] to be 54, got %v", cfg.SubTestConfig.IntArray[1])
	}

	if cfg.SubTestConfig.IntArray[2] != 55 {
		t.Fatalf("expected SubTestConfig.IntArray[2] to be 55, got %v", cfg.SubTestConfig.IntArray[2])
	}

	if len(cfg.SubTestConfig.Int8Array) != 3 {
		t.Fatalf("expected SubTestConfig.Int8Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int8Array))
	}

	if cfg.SubTestConfig.Int8Array[0] != 56 {
		t.Fatalf("expected SubTestConfig.Int8Array[0] to be 56, got %v", cfg.SubTestConfig.Int8Array[0])
	}

	if cfg.SubTestConfig.Int8Array[1] != 57 {
		t.Fatalf("expected SubTestConfig.Int8Array[1] to be 57, got %v", cfg.SubTestConfig.Int8Array[1])
	}

	if cfg.SubTestConfig.Int8Array[2] != 58 {
		t.Fatalf("expected SubTestConfig.Int8Array[2] to be 58, got %v", cfg.SubTestConfig.Int8Array[2])
	}

	if len(cfg.SubTestConfig.Int16Array) != 3 {
		t.Fatalf("expected SubTestConfig.Int16Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int16Array))
	}

	if cfg.SubTestConfig.Int16Array[0] != 59 {
		t.Fatalf("expected SubTestConfig.Int16Array[0] to be 59, got %v", cfg.SubTestConfig.Int16Array[0])
	}

	if cfg.SubTestConfig.Int16Array[1] != 60 {
		t.Fatalf("expected SubTestConfig.Int16Array[1] to be 60, got %v", cfg.SubTestConfig.Int16Array[1])
	}

	if cfg.SubTestConfig.Int16Array[2] != 61 {
		t.Fatalf("expected SubTestConfig.Int16Array[2] to be 61, got %v", cfg.SubTestConfig.Int16Array[2])
	}

	if len(cfg.SubTestConfig.Int32Array) != 3 {
		t.Fatalf("expected SubTestConfig.Int32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int32Array))
	}

	if cfg.SubTestConfig.Int32Array[0] != 62 {
		t.Fatalf("expected SubTestConfig.Int32Array[0] to be 62, got %v", cfg.SubTestConfig.Int32Array[0])
	}

	if cfg.SubTestConfig.Int32Array[1] != 63 {
		t.Fatalf("expected SubTestConfig.Int32Array[1] to be 63, got %v", cfg.SubTestConfig.Int32Array[1])
	}

	if cfg.SubTestConfig.Int32Array[2] != 64 {
		t.Fatalf("expected SubTestConfig.Int32Array[2] to be 64, got %v", cfg.SubTestConfig.Int32Array[2])
	}

	if len(cfg.SubTestConfig.Int64Array) != 3 {
		t.Fatalf("expected SubTestConfig.Int64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int64Array))
	}

	if cfg.SubTestConfig.Int64Array[0] != 65 {
		t.Fatalf("expected SubTestConfig.Int64Array[0] to be 65, got %v", cfg.SubTestConfig.Int64Array[0])
	}

	if cfg.SubTestConfig.Int64Array[1] != 66 {
		t.Fatalf("expected SubTestConfig.Int64Array[1] to be 66, got %v", cfg.SubTestConfig.Int64Array[1])
	}

	if cfg.SubTestConfig.Int64Array[2] != 67 {
		t.Fatalf("expected SubTestConfig.Int64Array[2] to be 67, got %v", cfg.SubTestConfig.Int64Array[2])
	}

	if len(cfg.SubTestConfig.UintArray) != 3 {
		t.Fatalf("expected SubTestConfig.UintArray to have 3 elements, got %d", len(cfg.SubTestConfig.UintArray))
	}

	if cfg.SubTestConfig.UintArray[0] != 68 {
		t.Fatalf("expected SubTestConfig.UintArray[0] to be 68, got %v", cfg.SubTestConfig.UintArray[0])
	}

	if cfg.SubTestConfig.UintArray[1] != 69 {
		t.Fatalf("expected SubTestConfig.UintArray[1] to be 69, got %v", cfg.SubTestConfig.UintArray[1])
	}

	if cfg.SubTestConfig.UintArray[2] != 70 {
		t.Fatalf("expected SubTestConfig.UintArray[2] to be 70, got %v", cfg.SubTestConfig.UintArray[2])
	}

	if len(cfg.SubTestConfig.Uint8Array) != 3 {
		t.Fatalf("expected SubTestConfig.Uint8Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint8Array))
	}

	if cfg.SubTestConfig.Uint8Array[0] != 71 {
		t.Fatalf("expected SubTestConfig.Uint8Array[0] to be 71, got %v", cfg.SubTestConfig.Uint8Array[0])
	}

	if cfg.SubTestConfig.Uint8Array[1] != 72 {
		t.Fatalf("expected SubTestConfig.Uint8Array[1] to be 72, got %v", cfg.SubTestConfig.Uint8Array[1])
	}

	if cfg.SubTestConfig.Uint8Array[2] != 73 {
		t.Fatalf("expected SubTestConfig.Uint8Array[2] to be 73, got %v", cfg.SubTestConfig.Uint8Array[2])
	}

	if len(cfg.SubTestConfig.Uint16Array) != 3 {
		t.Fatalf("expected SubTestConfig.Uint16Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint16Array))
	}

	if cfg.SubTestConfig.Uint16Array[0] != 74 {
		t.Fatalf("expected SubTestConfig.Uint16Array[0] to be 74, got %v", cfg.SubTestConfig.Uint16Array[0])
	}

	if cfg.SubTestConfig.Uint16Array[1] != 75 {
		t.Fatalf("expected SubTestConfig.Uint16Array[1] to be 75, got %v", cfg.SubTestConfig.Uint16Array[1])
	}

	if cfg.SubTestConfig.Uint16Array[2] != 76 {
		t.Fatalf("expected SubTestConfig.Uint16Array[2] to be 76, got %v", cfg.SubTestConfig.Uint16Array[2])
	}

	if len(cfg.SubTestConfig.Uint32Array) != 3 {
		t.Fatalf("expected SubTestConfig.Uint32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint32Array))
	}

	if cfg.SubTestConfig.Uint32Array[0] != 77 {
		t.Fatalf("expected SubTestConfig.Uint32Array[0] to be 77, got %v", cfg.SubTestConfig.Uint32Array[0])
	}

	if cfg.SubTestConfig.Uint32Array[1] != 78 {
		t.Fatalf("expected SubTestConfig.Uint32Array[1] to be 78, got %v", cfg.SubTestConfig.Uint32Array[1])
	}

	if cfg.SubTestConfig.Uint32Array[2] != 79 {
		t.Fatalf("expected SubTestConfig.Uint32Array[2] to be 79, got %v", cfg.SubTestConfig.Uint32Array[2])
	}

	if len(cfg.SubTestConfig.Uint64Array) != 3 {
		t.Fatalf("expected SubTestConfig.Uint64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint64Array))
	}

	if cfg.SubTestConfig.Uint64Array[0] != 80 {
		t.Fatalf("expected SubTestConfig.Uint64Array[0] to be 80, got %v", cfg.SubTestConfig.Uint64Array[0])
	}

	if cfg.SubTestConfig.Uint64Array[1] != 81 {
		t.Fatalf("expected SubTestConfig.Uint64Array[1] to be 81, got %v", cfg.SubTestConfig.Uint64Array[1])
	}

	if cfg.SubTestConfig.Uint64Array[2] != 82 {
		t.Fatalf("expected SubTestConfig.Uint64Array[2] to be 82, got %v", cfg.SubTestConfig.Uint64Array[2])
	}

	if len(cfg.SubTestConfig.Float32Array) != 3 {
		t.Fatalf("expected SubTestConfig.Float32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Float32Array))
	}

	if cfg.SubTestConfig.Float32Array[0] != 83.0 {
		t.Fatalf("expected SubTestConfig.Float32Array[0] to be 83.0, got %v", cfg.SubTestConfig.Float32Array[0])
	}

	if cfg.SubTestConfig.Float32Array[1] != 84.0 {
		t.Fatalf("expected SubTestConfig.Float32Array[1] to be 84.0, got %v", cfg.SubTestConfig.Float32Array[1])
	}

	if cfg.SubTestConfig.Float32Array[2] != 85.0 {
		t.Fatalf("expected SubTestConfig.Float32Array[2] to be 85.0, got %v", cfg.SubTestConfig.Float32Array[2])
	}

	if len(cfg.SubTestConfig.Float64Array) != 3 {
		t.Fatalf("expected SubTestConfig.Float64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Float64Array))
	}

	if cfg.SubTestConfig.Float64Array[0] != 86.0 {
		t.Fatalf("expected SubTestConfig.Float64Array[0] to be 86.0, got %v", cfg.SubTestConfig.Float64Array[0])
	}

	if cfg.SubTestConfig.Float64Array[1] != 87.0 {
		t.Fatalf("expected SubTestConfig.Float64Array[1] to be 87.0, got %v", cfg.SubTestConfig.Float64Array[1])
	}

	if cfg.SubTestConfig.Float64Array[2] != 88.0 {
		t.Fatalf("expected SubTestConfig.Float64Array[2] to be 88.0, got %v", cfg.SubTestConfig.Float64Array[2])
	}

	if len(cfg.SubTestConfig.InterfaceArray) != 2 {
		t.Fatalf("expected SubTestConfig.InterfaceArray to have 2 elements, got %d", len(cfg.SubTestConfig.InterfaceArray))
	}

	if cfg.SubTestConfig.InterfaceArray[0] != "5.0" {
		t.Fatalf("expected SubTestConfig.InterfaceArray[0] to be '5.0', got %v", cfg.SubTestConfig.InterfaceArray[0])
	}

	if cfg.SubTestConfig.InterfaceArray[1] != "8.0" {
		t.Fatalf("expected SubTestConfig.InterfaceArray[1] to be '8.0', got %v", cfg.SubTestConfig.InterfaceArray[1])
	}
}

//nolint:golint,gocyclo
func TestConfigulatorFlags(t *testing.T) {
	t.Parallel()

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	c := New[testConfig]()
	c.WithPFlags(flags, &PFlagOptions{
		Separator: "-",
	})

	err := flags.Parse(
		[]string{
			"--bool=false",
			"--int=20",
			"--int8=21",
			"--int16=22",
			"--int32=23",
			"--int64=24",
			"--uint=25",
			"--uint8=26",
			"--uint16=27",
			"--uint32=28",
			"--uint64=29",
			"--float32=30.0",
			"--float64=31.0",
			"--string=32.0",
			"--stringArray=53.0,54.0",
			"--boolArray=false,false,true",
			"--intArray=33,34,35",
			"--int8Array=36,37,38",
			"--int16Array=39,40,41",
			"--int32Array=42,43,44",
			"--int64Array=45,46,47",
			"--uintArray=48,49,50",
			"--uint8Array=51,52,53",
			"--uint16Array=54,55,56",
			"--uint32Array=57,58,59",
			"--uint64Array=60,61,62",
			"--float32Array=63.0,64.0,65.0",
			"--float64Array=66.0,67.0,68.0",
			"--subTestConfig-bool=false",
			"--subTestConfig-int=40",
			"--subTestConfig-int8=41",
			"--subTestConfig-int16=42",
			"--subTestConfig-int32=43",
			"--subTestConfig-int64=44",
			"--subTestConfig-uint=45",
			"--subTestConfig-uint8=46",
			"--subTestConfig-uint16=47",
			"--subTestConfig-uint32=48",
			"--subTestConfig-uint64=49",
			"--subTestConfig-float32=50.0",
			"--subTestConfig-float64=51.0",
			"--subTestConfig-string=52.0",
			"--subTestConfig-stringArray=68,69",
			"--subTestConfig-boolArray=false,false,true",
			"--subTestConfig-intArray=53,54,55",
			"--subTestConfig-int8Array=56",
			"--subTestConfig-int8Array=57",
			"--subTestConfig-int8Array=58",
			"--subTestConfig-int16Array=59,60,61",
			"--subTestConfig-int32Array=62,63,64",
			"--subTestConfig-int64Array=65,66,67",
			"--subTestConfig-uintArray=68,69,70",
			"--subTestConfig-uint8Array=71,72,73",
			"--subTestConfig-uint16Array=74,75,76",
			"--subTestConfig-uint32Array=77,78,79",
			"--subTestConfig-uint64Array=80,81,82",
			"--subTestConfig-float32Array=83.0,84.0,85.0",
			"--subTestConfig-float64Array=86.0,87.0,88.0",
		})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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

	if len(cfg.StringArray) != 2 {
		t.Fatalf("expected StringArray to have 2 elements, got %d", len(cfg.StringArray))
	}

	if cfg.StringArray[0] != "53.0" {
		t.Fatalf("expected StringArray[0] to be '53.0', got '%s'", cfg.StringArray[0])
	}

	if cfg.StringArray[1] != "54.0" {
		t.Fatalf("expected StringArray[1] to be '54.0', got '%s'", cfg.StringArray[1])
	}

	if len(cfg.IntArray) != 3 {
		t.Fatalf("expected IntArray to have 3 elements, got %d", len(cfg.IntArray))
	}

	if cfg.IntArray[0] != 33 {
		t.Fatalf("expected IntArray[0] to be 33, got %v", cfg.IntArray[0])
	}

	if cfg.IntArray[1] != 34 {
		t.Fatalf("expected IntArray[1] to be 34, got %v", cfg.IntArray[1])
	}

	if cfg.IntArray[2] != 35 {
		t.Fatalf("expected IntArray[2] to be 35, got %v", cfg.IntArray[2])
	}

	if len(cfg.Int8Array) != 3 {
		t.Fatalf("expected Int8Array to have 3 elements, got %d", len(cfg.Int8Array))
	}

	if cfg.Int8Array[0] != 36 {
		t.Fatalf("expected Int8Array[0] to be 36, got %v", cfg.Int8Array[0])
	}

	if cfg.Int8Array[1] != 37 {
		t.Fatalf("expected Int8Array[1] to be 37, got %v", cfg.Int8Array[1])
	}

	if cfg.Int8Array[2] != 38 {
		t.Fatalf("expected Int8Array[2] to be 38, got %v", cfg.Int8Array[2])
	}

	if len(cfg.Int16Array) != 3 {
		t.Fatalf("expected Int16Array to have 3 elements, got %d", len(cfg.Int16Array))
	}

	if cfg.Int16Array[0] != 39 {
		t.Fatalf("expected Int16Array[0] to be 39, got %v", cfg.Int16Array[0])
	}

	if cfg.Int16Array[1] != 40 {
		t.Fatalf("expected Int16Array[1] to be 40, got %v", cfg.Int16Array[1])
	}

	if cfg.Int16Array[2] != 41 {
		t.Fatalf("expected Int16Array[2] to be 41, got %v", cfg.Int16Array[2])
	}

	if len(cfg.Int32Array) != 3 {
		t.Fatalf("expected Int32Array to have 3 elements, got %d", len(cfg.Int32Array))
	}

	if cfg.Int32Array[0] != 42 {
		t.Fatalf("expected Int32Array[0] to be 42, got %v", cfg.Int32Array[0])
	}

	if cfg.Int32Array[1] != 43 {
		t.Fatalf("expected Int32Array[1] to be 43, got %v", cfg.Int32Array[1])
	}

	if cfg.Int32Array[2] != 44 {
		t.Fatalf("expected Int32Array[2] to be 44, got %v", cfg.Int32Array[2])
	}

	if len(cfg.Int64Array) != 3 {
		t.Fatalf("expected Int64Array to have 3 elements, got %d", len(cfg.Int64Array))
	}

	if cfg.Int64Array[0] != 45 {
		t.Fatalf("expected Int64Array[0] to be 45, got %v", cfg.Int64Array[0])
	}

	if cfg.Int64Array[1] != 46 {
		t.Fatalf("expected Int64Array[1] to be 46, got %v", cfg.Int64Array[1])
	}

	if cfg.Int64Array[2] != 47 {
		t.Fatalf("expected Int64Array[2] to be 47, got %v", cfg.Int64Array[2])
	}

	if len(cfg.UintArray) != 3 {
		t.Fatalf("expected UintArray to have 3 elements, got %d", len(cfg.UintArray))
	}

	if cfg.UintArray[0] != 48 {
		t.Fatalf("expected UintArray[0] to be 48, got %v", cfg.UintArray[0])
	}

	if cfg.UintArray[1] != 49 {
		t.Fatalf("expected UintArray[1] to be 49, got %v", cfg.UintArray[1])
	}

	if cfg.UintArray[2] != 50 {
		t.Fatalf("expected UintArray[2] to be 50, got %v", cfg.UintArray[2])
	}

	if len(cfg.Uint8Array) != 3 {
		t.Fatalf("expected Uint8Array to have 3 elements, got %d", len(cfg.Uint8Array))
	}

	if cfg.Uint8Array[0] != 51 {
		t.Fatalf("expected Uint8Array[0] to be 51, got %v", cfg.Uint8Array[0])
	}

	if cfg.Uint8Array[1] != 52 {
		t.Fatalf("expected Uint8Array[1] to be 52, got %v", cfg.Uint8Array[1])
	}

	if cfg.Uint8Array[2] != 53 {
		t.Fatalf("expected Uint8Array[2] to be 53, got %v", cfg.Uint8Array[2])
	}

	if len(cfg.Uint16Array) != 3 {
		t.Fatalf("expected Uint16Array to have 3 elements, got %d", len(cfg.Uint16Array))
	}

	if cfg.Uint16Array[0] != 54 {
		t.Fatalf("expected Uint16Array[0] to be 54, got %v", cfg.Uint16Array[0])
	}

	if cfg.Uint16Array[1] != 55 {
		t.Fatalf("expected Uint16Array[1] to be 55, got %v", cfg.Uint16Array[1])
	}

	if cfg.Uint16Array[2] != 56 {
		t.Fatalf("expected Uint16Array[2] to be 56, got %v", cfg.Uint16Array[2])
	}

	if len(cfg.Uint32Array) != 3 {
		t.Fatalf("expected Uint32Array to have 3 elements, got %d", len(cfg.Uint32Array))
	}

	if cfg.Uint32Array[0] != 57 {
		t.Fatalf("expected Uint32Array[0] to be 57, got %v", cfg.Uint32Array[0])
	}

	if cfg.Uint32Array[1] != 58 {
		t.Fatalf("expected Uint32Array[1] to be 58, got %v", cfg.Uint32Array[1])
	}

	if cfg.Uint32Array[2] != 59 {
		t.Fatalf("expected Uint32Array[2] to be 59, got %v", cfg.Uint32Array[2])
	}

	if len(cfg.Uint64Array) != 3 {
		t.Fatalf("expected Uint64Array to have 3 elements, got %d", len(cfg.Uint64Array))
	}

	if cfg.Uint64Array[0] != 60 {
		t.Fatalf("expected Uint64Array[0] to be 60, got %v", cfg.Uint64Array[0])
	}

	if cfg.Uint64Array[1] != 61 {
		t.Fatalf("expected Uint64Array[1] to be 61, got %v", cfg.Uint64Array[1])
	}

	if cfg.Uint64Array[2] != 62 {
		t.Fatalf("expected Uint64Array[2] to be 62, got %v", cfg.Uint64Array[2])
	}

	if len(cfg.Float32Array) != 3 {
		t.Fatalf("expected Float32Array to have 3 elements, got %d", len(cfg.Float32Array))
	}

	if cfg.Float32Array[0] != 63.0 {
		t.Fatalf("expected Float32Array[0] to be 63.0, got %v", cfg.Float32Array[0])
	}

	if cfg.Float32Array[1] != 64.0 {
		t.Fatalf("expected Float32Array[1] to be 64.0, got %v", cfg.Float32Array[1])
	}

	if cfg.Float32Array[2] != 65.0 {
		t.Fatalf("expected Float32Array[2] to be 65.0, got %v", cfg.Float32Array[2])
	}

	if len(cfg.Float64Array) != 3 {
		t.Fatalf("expected Float64Array to have 3 elements, got %d", len(cfg.Float64Array))
	}

	if cfg.Float64Array[0] != 66.0 {
		t.Fatalf("expected Float64Array[0] to be 66.0, got %v", cfg.Float64Array[0])
	}

	if cfg.Float64Array[1] != 67.0 {
		t.Fatalf("expected Float64Array[1] to be 67.0, got %v", cfg.Float64Array[1])
	}

	if cfg.Float64Array[2] != 68.0 {
		t.Fatalf("expected Float64Array[2] to be 68.0, got %v", cfg.Float64Array[2])
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

	if len(cfg.SubTestConfig.StringArray) != 2 {
		t.Fatalf("expected SubTestConfig.StringArray to have 2 elements, got %d", len(cfg.SubTestConfig.StringArray))
	}

	if cfg.SubTestConfig.StringArray[0] != "68" {
		t.Fatalf("expected SubTestConfig.StringArray[0] to be '68', got '%s'", cfg.SubTestConfig.StringArray[0])
	}

	if cfg.SubTestConfig.StringArray[1] != "69" {
		t.Fatalf("expected SubTestConfig.StringArray[1] to be '69', got '%s'", cfg.SubTestConfig.StringArray[1])
	}

	if len(cfg.SubTestConfig.BoolArray) != 3 {
		t.Fatalf("expected SubTestConfig.BoolArray to have 3 elements, got %d", len(cfg.SubTestConfig.BoolArray))
	}

	if cfg.SubTestConfig.BoolArray[0] != false {
		t.Fatalf("expected SubTestConfig.BoolArray[0] to be false, got %v", cfg.SubTestConfig.BoolArray[0])
	}

	if cfg.SubTestConfig.BoolArray[1] != false {
		t.Fatalf("expected SubTestConfig.BoolArray[1] to be false, got %v", cfg.SubTestConfig.BoolArray[1])
	}

	if cfg.SubTestConfig.BoolArray[2] != true {
		t.Fatalf("expected SubTestConfig.BoolArray[2] to be true, got %v", cfg.SubTestConfig.BoolArray[2])
	}

	if len(cfg.SubTestConfig.IntArray) != 3 {
		t.Fatalf("expected SubTestConfig.IntArray to have 3 elements, got %d", len(cfg.SubTestConfig.IntArray))
	}

	if cfg.SubTestConfig.IntArray[0] != 53 {
		t.Fatalf("expected SubTestConfig.IntArray[0] to be 53, got %v", cfg.SubTestConfig.IntArray[0])
	}

	if cfg.SubTestConfig.IntArray[1] != 54 {
		t.Fatalf("expected SubTestConfig.IntArray[1] to be 54, got %v", cfg.SubTestConfig.IntArray[1])
	}

	if cfg.SubTestConfig.IntArray[2] != 55 {
		t.Fatalf("expected SubTestConfig.IntArray[2] to be 55, got %v", cfg.SubTestConfig.IntArray[2])
	}

	if len(cfg.SubTestConfig.Int8Array) != 3 {
		t.Fatalf("expected SubTestConfig.Int8Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int8Array))
	}

	if cfg.SubTestConfig.Int8Array[0] != 56 {
		t.Fatalf("expected SubTestConfig.Int8Array[0] to be 56, got %v", cfg.SubTestConfig.Int8Array[0])
	}

	if cfg.SubTestConfig.Int8Array[1] != 57 {
		t.Fatalf("expected SubTestConfig.Int8Array[1] to be 57, got %v", cfg.SubTestConfig.Int8Array[1])
	}

	if cfg.SubTestConfig.Int8Array[2] != 58 {
		t.Fatalf("expected SubTestConfig.Int8Array[2] to be 58, got %v", cfg.SubTestConfig.Int8Array[2])
	}

	if len(cfg.SubTestConfig.Int16Array) != 3 {
		t.Fatalf("expected SubTestConfig.Int16Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int16Array))
	}

	if cfg.SubTestConfig.Int16Array[0] != 59 {
		t.Fatalf("expected SubTestConfig.Int16Array[0] to be 59, got %v", cfg.SubTestConfig.Int16Array[0])
	}

	if cfg.SubTestConfig.Int16Array[1] != 60 {
		t.Fatalf("expected SubTestConfig.Int16Array[1] to be 60, got %v", cfg.SubTestConfig.Int16Array[1])
	}

	if cfg.SubTestConfig.Int16Array[2] != 61 {
		t.Fatalf("expected SubTestConfig.Int16Array[2] to be 61, got %v", cfg.SubTestConfig.Int16Array[2])
	}

	if len(cfg.SubTestConfig.Int32Array) != 3 {
		t.Fatalf("expected SubTestConfig.Int32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int32Array))
	}

	if cfg.SubTestConfig.Int32Array[0] != 62 {
		t.Fatalf("expected SubTestConfig.Int32Array[0] to be 62, got %v", cfg.SubTestConfig.Int32Array[0])
	}

	if cfg.SubTestConfig.Int32Array[1] != 63 {
		t.Fatalf("expected SubTestConfig.Int32Array[1] to be 63, got %v", cfg.SubTestConfig.Int32Array[1])
	}

	if cfg.SubTestConfig.Int32Array[2] != 64 {
		t.Fatalf("expected SubTestConfig.Int32Array[2] to be 64, got %v", cfg.SubTestConfig.Int32Array[2])
	}

	if len(cfg.SubTestConfig.Int64Array) != 3 {
		t.Fatalf("expected SubTestConfig.Int64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Int64Array))
	}

	if cfg.SubTestConfig.Int64Array[0] != 65 {
		t.Fatalf("expected SubTestConfig.Int64Array[0] to be 65, got %v", cfg.SubTestConfig.Int64Array[0])
	}

	if cfg.SubTestConfig.Int64Array[1] != 66 {
		t.Fatalf("expected SubTestConfig.Int64Array[1] to be 66, got %v", cfg.SubTestConfig.Int64Array[1])
	}

	if cfg.SubTestConfig.Int64Array[2] != 67 {
		t.Fatalf("expected SubTestConfig.Int64Array[2] to be 67, got %v", cfg.SubTestConfig.Int64Array[2])
	}

	if len(cfg.SubTestConfig.UintArray) != 3 {
		t.Fatalf("expected SubTestConfig.UintArray to have 3 elements, got %d", len(cfg.SubTestConfig.UintArray))
	}

	if cfg.SubTestConfig.UintArray[0] != 68 {
		t.Fatalf("expected SubTestConfig.UintArray[0] to be 68, got %v", cfg.SubTestConfig.UintArray[0])
	}

	if cfg.SubTestConfig.UintArray[1] != 69 {
		t.Fatalf("expected SubTestConfig.UintArray[1] to be 69, got %v", cfg.SubTestConfig.UintArray[1])
	}

	if cfg.SubTestConfig.UintArray[2] != 70 {
		t.Fatalf("expected SubTestConfig.UintArray[2] to be 70, got %v", cfg.SubTestConfig.UintArray[2])
	}

	if len(cfg.SubTestConfig.Uint8Array) != 3 {
		t.Fatalf("expected SubTestConfig.Uint8Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint8Array))
	}

	if cfg.SubTestConfig.Uint8Array[0] != 71 {
		t.Fatalf("expected SubTestConfig.Uint8Array[0] to be 71, got %v", cfg.SubTestConfig.Uint8Array[0])
	}

	if cfg.SubTestConfig.Uint8Array[1] != 72 {
		t.Fatalf("expected SubTestConfig.Uint8Array[1] to be 72, got %v", cfg.SubTestConfig.Uint8Array[1])
	}

	if cfg.SubTestConfig.Uint8Array[2] != 73 {
		t.Fatalf("expected SubTestConfig.Uint8Array[2] to be 73, got %v", cfg.SubTestConfig.Uint8Array[2])
	}

	if len(cfg.SubTestConfig.Uint16Array) != 3 {
		t.Fatalf("expected SubTestConfig.Uint16Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint16Array))
	}

	if cfg.SubTestConfig.Uint16Array[0] != 74 {
		t.Fatalf("expected SubTestConfig.Uint16Array[0] to be 74, got %v", cfg.SubTestConfig.Uint16Array[0])
	}

	if cfg.SubTestConfig.Uint16Array[1] != 75 {
		t.Fatalf("expected SubTestConfig.Uint16Array[1] to be 75, got %v", cfg.SubTestConfig.Uint16Array[1])
	}

	if cfg.SubTestConfig.Uint16Array[2] != 76 {
		t.Fatalf("expected SubTestConfig.Uint16Array[2] to be 76, got %v", cfg.SubTestConfig.Uint16Array[2])
	}

	if len(cfg.SubTestConfig.Uint32Array) != 3 {
		t.Fatalf("expected SubTestConfig.Uint32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint32Array))
	}

	if cfg.SubTestConfig.Uint32Array[0] != 77 {
		t.Fatalf("expected SubTestConfig.Uint32Array[0] to be 77, got %v", cfg.SubTestConfig.Uint32Array[0])
	}

	if cfg.SubTestConfig.Uint32Array[1] != 78 {
		t.Fatalf("expected SubTestConfig.Uint32Array[1] to be 78, got %v", cfg.SubTestConfig.Uint32Array[1])
	}

	if cfg.SubTestConfig.Uint32Array[2] != 79 {
		t.Fatalf("expected SubTestConfig.Uint32Array[2] to be 79, got %v", cfg.SubTestConfig.Uint32Array[2])
	}

	if len(cfg.SubTestConfig.Uint64Array) != 3 {
		t.Fatalf("expected SubTestConfig.Uint64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Uint64Array))
	}

	if cfg.SubTestConfig.Uint64Array[0] != 80 {
		t.Fatalf("expected SubTestConfig.Uint64Array[0] to be 80, got %v", cfg.SubTestConfig.Uint64Array[0])
	}

	if cfg.SubTestConfig.Uint64Array[1] != 81 {
		t.Fatalf("expected SubTestConfig.Uint64Array[1] to be 81, got %v", cfg.SubTestConfig.Uint64Array[1])
	}

	if cfg.SubTestConfig.Uint64Array[2] != 82 {
		t.Fatalf("expected SubTestConfig.Uint64Array[2] to be 82, got %v", cfg.SubTestConfig.Uint64Array[2])
	}

	if len(cfg.SubTestConfig.Float32Array) != 3 {
		t.Fatalf("expected SubTestConfig.Float32Array to have 3 elements, got %d", len(cfg.SubTestConfig.Float32Array))
	}

	if cfg.SubTestConfig.Float32Array[0] != 83.0 {
		t.Fatalf("expected SubTestConfig.Float32Array[0] to be 83.0, got %v", cfg.SubTestConfig.Float32Array[0])
	}

	if cfg.SubTestConfig.Float32Array[1] != 84.0 {
		t.Fatalf("expected SubTestConfig.Float32Array[1] to be 84.0, got %v", cfg.SubTestConfig.Float32Array[1])
	}

	if cfg.SubTestConfig.Float32Array[2] != 85.0 {
		t.Fatalf("expected SubTestConfig.Float32Array[2] to be 85.0, got %v", cfg.SubTestConfig.Float32Array[2])
	}

	if len(cfg.SubTestConfig.Float64Array) != 3 {
		t.Fatalf("expected SubTestConfig.Float64Array to have 3 elements, got %d", len(cfg.SubTestConfig.Float64Array))
	}

	if cfg.SubTestConfig.Float64Array[0] != 86.0 {
		t.Fatalf("expected SubTestConfig.Float64Array[0] to be 86.0, got %v", cfg.SubTestConfig.Float64Array[0])
	}

	if cfg.SubTestConfig.Float64Array[1] != 87.0 {
		t.Fatalf("expected SubTestConfig.Float64Array[1] to be 87.0, got %v", cfg.SubTestConfig.Float64Array[1])
	}

	if cfg.SubTestConfig.Float64Array[2] != 88.0 {
		t.Fatalf("expected SubTestConfig.Float64Array[2] to be 88.0, got %v", cfg.SubTestConfig.Float64Array[2])
	}
}

func TestNonDefault(t *testing.T) {
	pflags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	c := New[nonDefaultsTestConfig]().
		WithPFlags(pflags, nil).
		WithEnvironmentVariables(&EnvironmentVariableOptions{
			Prefix:    "TEST_",
			Separator: "_",
		})

	cfg, err := c.Default()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Bool != false {
		t.Fatalf("expected Bool to be true, got %v", cfg.Bool)
	}

	if cfg.Int != 0 {
		t.Fatalf("expected Int to be 0, got %v", cfg.Int)
	}

	if cfg.Int8 != 0 {
		t.Fatalf("expected Int8 to be 0, got %v", cfg.Int8)
	}

	if cfg.Int16 != 0 {
		t.Fatalf("expected Int16 to be 0, got %v", cfg.Int16)
	}

	if cfg.Int32 != 0 {
		t.Fatalf("expected Int32 to be 0, got %v", cfg.Int32)
	}

	if cfg.Int64 != 0 {
		t.Fatalf("expected Int64 to be 0, got %v", cfg.Int64)
	}

	if cfg.Uint != 0 {
		t.Fatalf("expected Uint to be 0, got %v", cfg.Uint)
	}

	if cfg.Uint8 != 0 {
		t.Fatalf("expected Uint8 to be 0, got %v", cfg.Uint8)
	}

	if cfg.Uint16 != 0 {
		t.Fatalf("expected Uint16 to be 0, got %v", cfg.Uint16)
	}

	if cfg.Uint32 != 0 {
		t.Fatalf("expected Uint32 to be 0, got %v", cfg.Uint32)
	}

	if cfg.Uint64 != 0 {
		t.Fatalf("expected Uint64 to be 0, got %v", cfg.Uint64)
	}

	if cfg.Float32 != 0.0 {
		t.Fatalf("expected Float32 to be 0.0, got %v", cfg.Float32)
	}

	if cfg.Float64 != 0.0 {
		t.Fatalf("expected Float64 to be 0.0, got %v", cfg.Float64)
	}

	if cfg.String != "" {
		t.Fatalf("expected String to be empty, got '%s'", cfg.String)
	}

	if cfg.Any != nil {
		t.Fatalf("expected Any to be nil, got %v", cfg.Any)
	}

	if len(cfg.StringArray) != 0 {
		t.Fatalf("expected StringArray to be empty, got %d elements", len(cfg.StringArray))
	}

	if len(cfg.BoolArray) != 0 {
		t.Fatalf("expected BoolArray to be empty, got %d elements", len(cfg.BoolArray))
	}

	if len(cfg.IntArray) != 0 {
		t.Fatalf("expected IntArray to be empty, got %d elements", len(cfg.IntArray))
	}

	if len(cfg.Int8Array) != 0 {
		t.Fatalf("expected Int8Array to be empty, got %d elements", len(cfg.Int8Array))
	}

	if len(cfg.Int16Array) != 0 {
		t.Fatalf("expected Int16Array to be empty, got %d elements", len(cfg.Int16Array))
	}

	if len(cfg.Int32Array) != 0 {
		t.Fatalf("expected Int32Array to be empty, got %d elements", len(cfg.Int32Array))
	}

	if len(cfg.Int64Array) != 0 {
		t.Fatalf("expected Int64Array to be empty, got %d elements", len(cfg.Int64Array))
	}

	if len(cfg.UintArray) != 0 {
		t.Fatalf("expected UintArray to be empty, got %d elements", len(cfg.UintArray))
	}

	if len(cfg.Uint8Array) != 0 {
		t.Fatalf("expected Uint8Array to be empty, got %d elements", len(cfg.Uint8Array))
	}

	if len(cfg.Uint16Array) != 0 {
		t.Fatalf("expected Uint16Array to be empty, got %d elements", len(cfg.Uint16Array))
	}

	if len(cfg.Uint32Array) != 0 {
		t.Fatalf("expected Uint32Array to be empty, got %d elements", len(cfg.Uint32Array))
	}

	if len(cfg.Uint64Array) != 0 {
		t.Fatalf("expected Uint64Array to be empty, got %d elements", len(cfg.Uint64Array))
	}

	if len(cfg.Float32Array) != 0 {
		t.Fatalf("expected Float32Array to be empty, got %d elements", len(cfg.Float32Array))
	}

	if len(cfg.Float64Array) != 0 {
		t.Fatalf("expected Float64Array to be empty, got %d elements", len(cfg.Float64Array))
	}

	if len(cfg.InterfaceArray) != 0 {
		t.Fatalf("expected InterfaceArray to be empty, got %d elements", len(cfg.InterfaceArray))
	}

	if cfg.SubTestConfig.Bool != false {
		t.Fatalf("expected SubTestConfig.Bool to be false, got %v", cfg.SubTestConfig.Bool)
	}

	if cfg.SubTestConfig.Int != 0 {
		t.Fatalf("expected SubTestConfig.Int to be 0, got %v", cfg.SubTestConfig.Int)
	}

	if cfg.SubTestConfig.Int8 != 0 {
		t.Fatalf("expected SubTestConfig.Int8 to be 0, got %v", cfg.SubTestConfig.Int8)
	}

	if cfg.SubTestConfig.Int16 != 0 {
		t.Fatalf("expected SubTestConfig.Int16 to be 0, got %v", cfg.SubTestConfig.Int16)
	}

	if cfg.SubTestConfig.Int32 != 0 {
		t.Fatalf("expected SubTestConfig.Int32 to be 0, got %v", cfg.SubTestConfig.Int32)
	}

	if cfg.SubTestConfig.Int64 != 0 {
		t.Fatalf("expected SubTestConfig.Int64 to be 0, got %v", cfg.SubTestConfig.Int64)
	}

	if cfg.SubTestConfig.Uint != 0 {
		t.Fatalf("expected SubTestConfig.Uint to be 0, got %v", cfg.SubTestConfig.Uint)
	}

	if cfg.SubTestConfig.Uint8 != 0 {
		t.Fatalf("expected SubTestConfig.Uint8 to be 0, got %v", cfg.SubTestConfig.Uint8)
	}

	if cfg.SubTestConfig.Uint16 != 0 {
		t.Fatalf("expected SubTestConfig.Uint16 to be 0, got %v", cfg.SubTestConfig.Uint16)
	}

	if cfg.SubTestConfig.Uint32 != 0 {
		t.Fatalf("expected SubTestConfig.Uint32 to be 0, got %v", cfg.SubTestConfig.Uint32)
	}

	if cfg.SubTestConfig.Uint64 != 0 {
		t.Fatalf("expected SubTestConfig.Uint64 to be 0, got %v", cfg.SubTestConfig.Uint64)
	}

	if cfg.SubTestConfig.Float32 != 0.0 {
		t.Fatalf("expected SubTestConfig.Float32 to be 0.0, got %v", cfg.SubTestConfig.Float32)
	}

	if cfg.SubTestConfig.Float64 != 0.0 {
		t.Fatalf("expected SubTestConfig.Float64 to be 0.0, got %v", cfg.SubTestConfig.Float64)
	}

	if cfg.SubTestConfig.String != "" {
		t.Fatalf("expected SubTestConfig.String to be empty, got '%s'", cfg.SubTestConfig.String)
	}

	if cfg.SubTestConfig.Any != nil {
		t.Fatalf("expected SubTestConfig.Any to be nil, got %v", cfg.SubTestConfig.Any)
	}

	if len(cfg.SubTestConfig.StringArray) != 0 {
		t.Fatalf("expected SubTestConfig.StringArray to be empty, got %d elements", len(cfg.SubTestConfig.StringArray))
	}

	if len(cfg.SubTestConfig.BoolArray) != 0 {
		t.Fatalf("expected SubTestConfig.BoolArray to be empty, got %d elements", len(cfg.SubTestConfig.BoolArray))
	}

	if len(cfg.SubTestConfig.IntArray) != 0 {
		t.Fatalf("expected SubTestConfig.IntArray to be empty, got %d elements", len(cfg.SubTestConfig.IntArray))
	}

	if len(cfg.SubTestConfig.Int8Array) != 0 {
		t.Fatalf("expected SubTestConfig.Int8Array to be empty, got %d elements", len(cfg.SubTestConfig.Int8Array))
	}

	if len(cfg.SubTestConfig.Int16Array) != 0 {
		t.Fatalf("expected SubTestConfig.Int16Array to be empty, got %d elements", len(cfg.SubTestConfig.Int16Array))
	}

	if len(cfg.SubTestConfig.Int32Array) != 0 {
		t.Fatalf("expected SubTestConfig.Int32Array to be empty, got %d elements", len(cfg.SubTestConfig.Int32Array))
	}

	if len(cfg.SubTestConfig.Int64Array) != 0 {
		t.Fatalf("expected SubTestConfig.Int64Array to be empty, got %d elements", len(cfg.SubTestConfig.Int64Array))
	}

	if len(cfg.SubTestConfig.UintArray) != 0 {
		t.Fatalf("expected SubTestConfig.UintArray to be empty, got %d elements", len(cfg.SubTestConfig.UintArray))
	}

	if len(cfg.SubTestConfig.Uint8Array) != 0 {
		t.Fatalf("expected SubTestConfig.Uint8Array to be empty, got %d elements", len(cfg.SubTestConfig.Uint8Array))
	}

	if len(cfg.SubTestConfig.Uint16Array) != 0 {
		t.Fatalf("expected SubTestConfig.Uint16Array to be empty, got %d elements", len(cfg.SubTestConfig.Uint16Array))
	}

	if len(cfg.SubTestConfig.Uint32Array) != 0 {
		t.Fatalf("expected SubTestConfig.Uint32Array to be empty, got %d elements", len(cfg.SubTestConfig.Uint32Array))
	}

	if len(cfg.SubTestConfig.Uint64Array) != 0 {
		t.Fatalf("expected SubTestConfig.Uint64Array to be empty, got %d elements", len(cfg.SubTestConfig.Uint64Array))
	}

	if len(cfg.SubTestConfig.Float32Array) != 0 {
		t.Fatalf("expected SubTestConfig.Float32Array to be empty, got %d elements", len(cfg.SubTestConfig.Float32Array))
	}

	if len(cfg.SubTestConfig.Float64Array) != 0 {
		t.Fatalf("expected SubTestConfig.Float64Array to be empty, got %d elements", len(cfg.SubTestConfig.Float64Array))
	}

	if len(cfg.SubTestConfig.InterfaceArray) != 0 {
		t.Fatalf("expected SubTestConfig.InterfaceArray to be empty, got %d elements", len(cfg.SubTestConfig.InterfaceArray))
	}
}
