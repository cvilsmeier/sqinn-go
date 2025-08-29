package prebuilt

import (
	_ "embed"
)

//go:embed "darwin-arm64.gz"
var gzipData []byte
