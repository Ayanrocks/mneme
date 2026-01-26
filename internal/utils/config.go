package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/pelletier/go-toml/v2"

	"mneme/internal/core"
	"mneme/internal/logger"
)

// PrettyPrintConfig displays the configuration in an intuitive bracket/tree format
// that's easy to read without TOML knowledge
func PrettyPrintConfig(configBytes []byte) error {
	var config core.DefaultConfig
	if err := toml.Unmarshal(configBytes, &config); err != nil {
		logger.Errorf("Failed to parse config: %+v", err)
		return err
	}

	// Create color styles
	sectionColor := color.New(color.FgCyan, color.Bold)
	keyColor := color.New(color.FgWhite)
	valueColor := color.New(color.FgGreen)
	arrayColor := color.New(color.FgYellow)
	checkmarkColor := color.New(color.FgGreen, color.Bold)

	// Helper function to print a checkmark for booleans
	printBool := func(value bool) string {
		if value {
			return checkmarkColor.Sprint("✓")
		}
		return color.New(color.FgRed).Sprint("✗")
	}

	// Helper function to print an array on multiple lines
	printArray := func(arr []string, indent string) string {
		if len(arr) == 0 {
			return arrayColor.Sprint("[]")
		}
		if len(arr) == 1 {
			return arrayColor.Sprintf("[\n%s    %q\n%s]", indent, arr[0], indent)
		}
		items := make([]string, len(arr))
		for i, item := range arr {
			if i == len(arr)-1 {
				items[i] = indent + "    " + "\"" + item + "\""
			} else {
				items[i] = indent + "    " + "\"" + item + "\","
			}
		}
		return arrayColor.Sprintf("[\n%s\n%s]", strings.Join(items, "\n"), indent)
	}

	// Helper function to print a value based on its type
	printValue := func(value reflect.Value, indent string) string {
		switch value.Kind() {
		case reflect.Bool:
			return printBool(value.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return valueColor.Sprint(value.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return valueColor.Sprint(value.Uint())
		case reflect.Float32, reflect.Float64:
			return valueColor.Sprint(value.Float())
		case reflect.String:
			return valueColor.Sprintf("%q", value.String())
		case reflect.Slice:
			if value.Type().Elem().Kind() == reflect.String {
				arr := value.Interface().([]string)
				return printArray(arr, indent)
			}
			return valueColor.Sprint(value.Interface())
		default:
			return valueColor.Sprint(value.Interface())
		}
	}

	// Helper function to print a struct section
	printSection := func(sectionName string, sectionValue reflect.Value) {
		logger.PrintRaw("%s", sectionColor.Sprint(sectionName))

		// Iterate through fields of the struct
		for i := 0; i < sectionValue.NumField(); i++ {
			field := sectionValue.Type().Field(i)
			fieldValue := sectionValue.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Get the toml tag name
			tomlTag := field.Tag.Get("toml")
			if tomlTag == "" {
				tomlTag = strings.ToLower(field.Name)
			}

			// Print the field
			logger.PrintRaw("%s", "  "+keyColor.Sprint(tomlTag)+": "+printValue(fieldValue, "  "))
		}
		logger.PrintRaw("")
	}

	// Print the config in bracket format
	logger.PrintRaw("")
	printSection("[INDEX]", reflect.ValueOf(config.Index))
	printSection("[SOURCES]", reflect.ValueOf(config.Sources))
	printSection("[WATCHER]", reflect.ValueOf(config.Watcher))
	printSection("[SEARCH]", reflect.ValueOf(config.Search))
	printSection("[RANKING]", reflect.ValueOf(config.Ranking))
	printSection("[LOGGING]", reflect.ValueOf(config.Logging))

	return nil
}

func ExpandFilePath(path string) (string, error) {
	// Handle tilde expansion first
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Errorf("Error getting user home directory: %+v", err)
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		// Join home directory with the rest of the path (after ~)
		// This handles cases like "~", "~/", "~/.config", etc.
		return filepath.Abs(filepath.Join(home, path[1:]))
	}

	// Handle relative paths (., ../, ../../../, etc.)
	// Convert to absolute path
	return filepath.Abs(path)
}
