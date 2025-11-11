package assets

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
)

//go:generate echo "embedding default sources"

//go:embed default-sources.yaml
var defaultSources []byte

// WriteDefaultSourcesIfMissing writes sources.yaml to targetDir if it does not exist.
func WriteDefaultSourcesIfMissing(targetDir string) error {
	if targetDir == "" {
		return errors.New("empty targetDir")
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(targetDir, "sources.yaml")
	if _, err := os.Stat(p); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.WriteFile(p, defaultSources, 0o644)
}
