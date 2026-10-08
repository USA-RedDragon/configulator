package configulator

import "testing"

type dashedSection struct {
	Enabled         bool `name:"enabled" description:"enabled"`
	IntervalSeconds int  `name:"interval-seconds" default:"60" description:"interval"`
}

type dashedConfig struct {
	PublicFrames dashedSection `name:"public-frames" description:"a section whose tag has a dash"`
}

func (dashedConfig) Validate() error { return nil }

// A nested struct whose tag has a dash is named by its Go field name in the
// environment (PUBLICFRAMES_), and that name must find the field again.
func TestNestedStructWithDashedTagFromEnvironment(t *testing.T) {
	t.Setenv("TEST_PUBLICFRAMES_ENABLED", "true")
	t.Setenv("TEST_PUBLICFRAMES_INTERVAL_SECONDS", "15")

	c := New[dashedConfig]()
	c.WithEnvironmentVariables(&EnvironmentVariableOptions{
		Prefix:    "TEST_",
		Separator: "_",
	})
	cfg, err := c.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.PublicFrames.Enabled {
		t.Error("PublicFrames.Enabled = false, want true")
	}
	if cfg.PublicFrames.IntervalSeconds != 15 {
		t.Errorf("PublicFrames.IntervalSeconds = %d, want 15", cfg.PublicFrames.IntervalSeconds)
	}
}
