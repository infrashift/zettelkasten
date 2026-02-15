package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"gopkg.in/yaml.v3"
)

//go:embed config-schema.cue
var configSchema string

//go:embed zettel-schema.cue
var zettelSchema string

// Config represents the application configuration.
type Config struct {
	RootPath  string  `json:"root_path" yaml:"root_path"`
	IndexPath string  `json:"index_path" yaml:"index_path"`
	GraphPath string  `json:"graph_path" yaml:"graph_path"`
	TodosPath string  `json:"todos_path" yaml:"todos_path"`
	Editor    string  `json:"editor" yaml:"editor"`
	Folders   Folders `json:"folders" yaml:"folders"`
}

// Folders represents the folder configuration.
type Folders struct {
	Fleeting  string `json:"fleeting" yaml:"fleeting"`
	Permanent string `json:"permanent" yaml:"permanent"`
	Tmp       string `json:"tmp" yaml:"tmp"`
}

// LoadSchema compiles and returns the #Config CUE definition.
func LoadSchema() (cue.Value, error) {
	ctx := cuecontext.New()
	val := ctx.CompileString(configSchema)
	if val.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to compile config schema: %w", val.Err())
	}
	return val.LookupPath(cue.ParsePath("#Config")), nil
}

// LoadZettelSchema compiles and returns the #Zettel CUE definition.
func LoadZettelSchema() (cue.Value, error) {
	ctx := cuecontext.New()
	val := ctx.CompileString(zettelSchema)
	if val.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to compile zettel schema: %w", val.Err())
	}
	return val.LookupPath(cue.ParsePath("#Zettel")), nil
}

// Load loads the configuration from the specified path.
// If configPath is empty, it returns the default configuration.
func Load(configPath string) (*Config, error) {
	schema, err := LoadSchema()
	if err != nil {
		return nil, err
	}

	ctx := schema.Context()

	// Start with schema defaults
	val := schema

	// If a config file is specified, load and unify it
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
			// File doesn't exist, use defaults
		} else {
			// Parse user config as CUE
			userVal := ctx.CompileBytes(data)
			if userVal.Err() != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", userVal.Err())
			}

			// Unify with schema
			val = schema.Unify(userVal)
			if val.Err() != nil {
				return nil, fmt.Errorf("config validation failed: %w", val.Err())
			}
		}
	}

	var cfg Config
	if err := val.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Expand ~ in root_path
	if len(cfg.RootPath) > 0 && cfg.RootPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.RootPath = filepath.Join(home, cfg.RootPath[1:])
	}

	return &cfg, nil
}

// ValidateZettelYAML validates YAML frontmatter against the #Zettel schema.
func ValidateZettelYAML(frontmatter []byte) error {
	schema, err := LoadZettelSchema()
	if err != nil {
		return err
	}

	// Parse YAML to map
	var data map[string]interface{}
	if err := yaml.Unmarshal(frontmatter, &data); err != nil {
		return fmt.Errorf("failed to parse frontmatter YAML: %w", err)
	}

	// Convert to CUE and validate
	ctx := schema.Context()
	dataVal := ctx.Encode(data)
	if dataVal.Err() != nil {
		return fmt.Errorf("failed to encode frontmatter: %w", dataVal.Err())
	}

	unified := schema.Unify(dataVal)
	// Use Concrete(true) to ensure all required fields have concrete values
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("frontmatter validation failed: %w", err)
	}

	return nil
}
