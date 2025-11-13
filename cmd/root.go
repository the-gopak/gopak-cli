package cmd

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/gopak/gopak-cli/internal/assets"
	"github.com/gopak/gopak-cli/internal/config"
	"github.com/gopak/gopak-cli/internal/logging"
)

var cfgFile string
var verbose bool
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "gopak",
	Short: "Universal Linux Installer",
}

func Execute() error { return rootCmd.Execute() }

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to any YAML file inside the config directory (default dir: ~/.config/gopak); all *.yaml in that directory are merged")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show detailed steps and commands")
	rootCmd.Version = version
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	var cfgDir string
	if cfgFile != "" {
		cfgDir = filepath.Dir(cfgFile)
	} else {
		dir, _ := os.UserConfigDir()
		if su := os.Getenv("SUDO_USER"); su != "" {
			if u, err := user.Lookup(su); err == nil && u.HomeDir != "" {
				dir = filepath.Join(u.HomeDir, ".config")
			}
		}
		cfgDir = dir + "/unilin"
	}
	// Ensure config directory and default sources.yaml exist
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = assets.WriteDefaultSourcesIfMissing(cfgDir)
	// Gather all YAML files and load
	entries, _ := os.ReadDir(cfgDir)
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		low := strings.ToLower(name)
		if strings.HasSuffix(low, ".yaml") || strings.HasSuffix(low, ".yml") {
			files = append(files, filepath.Join(cfgDir, name))
		}
	}
	if len(files) == 0 {
		logging.Error("no YAML config files found in " + cfgDir)
		os.Exit(1)
	}
	cfg, err := config.LoadFromFiles(files)
	if err != nil {
		logging.Error("config error: " + err.Error())
		os.Exit(1)
	}
	if err := config.ValidateAgainstSchema(cfg); err != nil {
		logging.Error("schema error: " + err.Error())
		os.Exit(1)
	}
	logging.Init()
	logging.SetVerbose(verbose)
}
