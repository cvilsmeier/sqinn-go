package sqinn_test

import (
	"io"
	"log"
	"testing"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func TestStdLogger(t *testing.T) {
	var l sqinn.StdLogger
	l.Log("foo")
	// Output:
	// foo2
	l.Logger = log.New(io.Discard, "", log.LstdFlags)
	l.Log("foo")
	// Output:
}

func TestNoLogger(t *testing.T) {
	var l sqinn.NoLogger
	l.Log("foo")
	// Output:
}
