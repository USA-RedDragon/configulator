package configulator

import (
	"context"
	"testing"

	"github.com/spf13/pflag"
)

type testConfig struct {
	Bool bool `config:"bool"`
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
	if subCtx.Value(configKey) == nil {
		t.Fatal("expected ConfigKey to be set in context")
	}
	if subCtx.Value(configKey) != c {
		t.Fatal("expected context to contain Configulator")
	}
}
