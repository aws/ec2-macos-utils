package system

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
)

// defaultConfigPath is an optional external cleanup-state configuration. When
// present it replaces the built-in default state list entirely.
const defaultConfigPath = "/usr/local/aws/ec2-macos-utils/cleanup-state.toml"

// defaultConfig is the built-in cleanup-state configuration used when no
// external config is present.
//
//go:embed cleanup-state.toml
var defaultConfig []byte

// stateConfig is the on-disk schema for cleanup-state configuration.
type stateConfig struct {
	State []StateEntry `toml:"state"`
}

// loadStateEntries returns the state entries from the external config at path
// if it exists, otherwise from the embedded default. A config that cannot be
// read or parsed is an error so a broken config never silently skips cleanup.
func loadStateEntries(path string) ([]StateEntry, error) {
	data := defaultConfig
	source := "embedded default config"

	if b, err := os.ReadFile(path); err == nil {
		data = b
		source = path
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("read cleanup-state config %q: %w", path, err)
	}

	var cfg stateConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse cleanup-state config (%s): %w", source, err)
	}

	return cfg.State, nil
}
