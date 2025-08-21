package prebuilt

import (
	_ "embed"
)

var sqinnName string = "sqinn.exe"

//go:embed "windows-amd64.gz"
var gzipData []byte
