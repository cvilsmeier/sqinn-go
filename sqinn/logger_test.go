package sqinn_test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func TestStdLogger(t *testing.T) {
	var l sqinn.StdLogger
	l.Log("foo")
	// Output:
	// foo
	l.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
	l.Log("foo")
	// Output:
}

func TestNoLogger(t *testing.T) {
	var l sqinn.NoLogger
	l.Log("foo")
	// Output:
}
