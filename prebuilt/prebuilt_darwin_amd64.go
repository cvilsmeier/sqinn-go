package prebuilt

import (
	_ "embed"
)

//go:embed "darwin-amd64.gz"
var gzipData []byte
