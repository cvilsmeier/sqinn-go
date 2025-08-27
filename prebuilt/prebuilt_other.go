//go:build !((linux || windows) && amd64) && !(linux && arm)

package prebuilt

var sqinnName string = "sqinn"
var gzipData []byte = nil
