//go:build !(linux && amd64) && !(windows && amd64)

package prebuilt

var sqinnName string = "sqinn"
var gzipData []byte = nil
