package prebuilt

import (
	_ "embed"
)

//go:embed "linux-amd64.gz"
var gzipData []byte
