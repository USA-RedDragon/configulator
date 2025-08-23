package configulator

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"reflect"

	inref "github.com/USA-RedDragon/configulator/internal/reflect"
	"github.com/USA-RedDragon/configulator/internal/wrapper"
	"gopkg.in/yaml.v3"
)

var (
	ErrConfigFileNotFound = fmt.Errorf("config file not found")
)

func (c *Configulator[C]) loadFromFile() error {
	configFile := make(map[string]any)
	for _, path := range c.fileOptions.Paths {
		stat, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to stat file: %w", err)
		}
		if stat.IsDir() {
			return fmt.Errorf("path is a directory")
		}
		fileReader, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer fileReader.Close()
		reader := bufio.NewReader(fileReader)
		dec := yaml.NewDecoder(reader)
		err = dec.Decode(&configFile)
		if err != nil {
			return fmt.Errorf("failed to decode yaml: %w", err)
		}
		break
	}

	if len(configFile) == 0 {
		return ErrConfigFileNotFound
	}

	return c.loadMap(configFile, []string{})
}

func (c *Configulator[C]) loadMap(configFile map[string]any, nest []string) error {
	refVal := reflect.ValueOf(c.cfg).Elem()
	for key, val := range configFile {
		// Split the key by the separator to get the nested structure
		nest := append(nest, key)

		_, err := inref.GetNestedStructFieldTypeByName(reflect.TypeOf(*c.cfg), nest, c.arraySeparator)
		if err != nil {
			continue
		}

		switch val := val.(type) {
		case map[string]any:
			err := c.loadMap(val, nest)
			if err != nil {
				slog.Error("Failed to load nested map", "key", key, "error", err)
				return fmt.Errorf("failed to load nested map for key %s: %w", key, err)
			}
			continue
		}

		wr := wrapper.WrappedValue{Value: val}
		err = inref.SetNestedStructValue(&refVal, nest, wr, c.arraySeparator)
		if err != nil {
			return fmt.Errorf("failed to set value for key %s: %w", key, err)
		}
	}

	return nil
}
