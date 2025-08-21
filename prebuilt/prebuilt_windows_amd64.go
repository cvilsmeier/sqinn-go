package prebuilt

import (
	_ "embed"
)

//go:embed "windows-amd64.gz"
var gzipData []byte
