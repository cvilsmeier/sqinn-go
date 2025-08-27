//go:build linux && arm

package prebuilt

import (
	_ "embed"
)

var sqinnName string = "sqinn"

//go:embed "linux-arm32.gz"
var gzipData []byte
