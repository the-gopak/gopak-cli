package manager

import (
	"fmt"
	"strings"

	"github.com/gopak/gopak-cli/internal/config"
)

const (
	placeholderPackage     = "{package}"
	placeholderPackageList = "{package_list}"
)

// expandCommandForNames expands a command for the given package names.
//
// Rules:
// - If command contains {package}: it is expanded per-package (group=false).
// - If command contains {package_list}: it is expanded once with all packages (group=true).
// - If command contains neither placeholder: it is executed once as-is (group=true).
// - If command contains both placeholders: returns an error.
func expandCommandForNames(cmd config.Command, names []string) (group bool, expanded []config.Command, err error) {
	if cmd.Command == "" {
		return true, nil, nil
	}
	if len(names) == 0 {
		return true, nil, fmt.Errorf("no package names provided")
	}

	hasPkg := strings.Contains(cmd.Command, placeholderPackage)
	hasList := strings.Contains(cmd.Command, placeholderPackageList)
	if hasPkg && hasList {
		return true, nil, fmt.Errorf("command contains both %s and %s", placeholderPackage, placeholderPackageList)
	}

	if hasPkg {
		out := make([]config.Command, 0, len(names))
		for _, n := range names {
			out = append(out, config.Command{
				Command:     strings.ReplaceAll(cmd.Command, placeholderPackage, n),
				RequireRoot: cmd.RequireRoot,
			})
		}
		return false, out, nil
	}

	cmdStr := cmd.Command
	if hasList {
		cmdStr = strings.ReplaceAll(cmdStr, placeholderPackageList, strings.Join(names, " "))
	}
	return true, []config.Command{{Command: cmdStr, RequireRoot: cmd.RequireRoot}}, nil
}

func expandCommandForName(cmd config.Command, name string) (config.Command, error) {
	_, expanded, err := expandCommandForNames(cmd, []string{name})
	if err != nil {
		return config.Command{}, err
	}
	if len(expanded) == 0 {
		return config.Command{}, nil
	}
	return expanded[0], nil
}
