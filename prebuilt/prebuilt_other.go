//go:build !(linux && amd64) && !(windows && amd64) && !(darwin && amd64) && !(darwin && arm64)

package prebuilt

var gzipData []byte = nil
