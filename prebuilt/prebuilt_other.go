//go:build !((linux || windows) && amd64)

package prebuilt

var sqinnName string = "sqinn"
var gzipData []byte = nil
