package prebuilt

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func Extract() (_dirname string, _filename string, _err error) {
	if len(gzipData) == 0 {
		platform := runtime.GOOS + "_" + runtime.GOARCH
		return "", "", fmt.Errorf("no embedded prebuilt sqinn binary found for %s, please build your own, see https://github.com/cvilsmeier/sqinn for build instructions", platform)
	}
	tempdir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", "", err
	}
	exeName := "sqinn"
	if runtime.GOOS == "windows" {
		exeName = "sqinn.exe"
	}
	tempname := filepath.Join(tempdir, exeName)
	gr, err := gzip.NewReader(bytes.NewReader(gzipData))
	if err != nil {
		os.RemoveAll(tempdir)
		return "", "", err
	}
	f, err := os.OpenFile(tempname, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		os.RemoveAll(tempdir)
		return "", "", err
	}
	defer f.Close()
	_, err = io.Copy(f, gr)
	if err != nil {
		os.RemoveAll(tempdir)
		return "", "", err
	}
	return tempdir, tempname, nil
}
