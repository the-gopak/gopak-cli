package manager

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type execCache struct {
	Entries map[string]time.Time `json:"entries"`
}

func execCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "gopak")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".cache", "gopak")
	}
	return filepath.Join(home, ".cache", "gopak")
}

func execCachePath() string {
	return filepath.Join(execCacheDir(), "exec-cache.json")
}

func execLockPath(pkg string) string {
	return filepath.Join(execCacheDir(), "exec-"+pkg+".lock")
}

func loadExecCache() execCache {
	c := execCache{Entries: map[string]time.Time{}}
	data, err := os.ReadFile(execCachePath())
	if err != nil {
		return c
	}
	_ = json.Unmarshal(data, &c)
	if c.Entries == nil {
		c.Entries = map[string]time.Time{}
	}
	return c
}

func saveExecCache(c execCache) {
	dir := execCacheDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	_ = os.WriteFile(execCachePath(), data, 0o600)
}

func (c *execCache) IsFresh(name string, ttl time.Duration) bool {
	t, ok := c.Entries[name]
	if !ok {
		return false
	}
	return time.Since(t) < ttl
}

func (c *execCache) Touch(name string) {
	c.Entries[name] = time.Now()
}

