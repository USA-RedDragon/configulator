package configulator

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
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

	slog.Info("Loaded config file", "config", configFile)
	return c.loadMap(configFile)
}

func (c *Configulator[C]) loadMap(configFile map[string]any) error {
	// typ := reflect.TypeOf(c.cfg).Elem()
	// if typ.Kind() != reflect.Struct {
	// 	return fmt.Errorf("expected struct, got %v", typ.Kind())
	// }

	// for i := range typ.NumField() {
	// 	field := typ.Field(i)
	// 	if field.Type.Kind() == reflect.Struct {

	// 	} else if tag := field.Tag.Get("config"); tag != "" {
	// 		tagInfo, err := tags.ExtractStructTag(field, tag, c.arraySeparator)
	// 		if err != nil {
	// 			return err
	// 		}

	// 	}
	// }
	return nil
}
