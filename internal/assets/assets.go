package assets

import (
	_ "embed"
)

//go:generate echo "embedding default sources"

//go:embed default-sources.yaml
var DefaultSources []byte
