package prebuilt

import (
	_ "embed"
)

var sqinnName string = "sqinn"

//go:embed "linux-amd64.gz"
var gzipData []byte
