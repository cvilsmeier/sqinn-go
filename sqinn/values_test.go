package sqinn_test

import (
	"testing"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func TestAnyValue(t *testing.T) {
	any := sqinn.AnyValue{
		Int: sqinn.IntValue{
			Set:   true,
			Value: 13,
		},
	}
	assert(t, any.AsInt() == 13, "wrong %v", any.AsInt())
	assert(t, any.AsInt64() == 0, "wrong %v", any.AsInt64())
	assert(t, any.AsDouble() == 0, "wrong %v", any.AsDouble())
	assert(t, any.AsString() == "", "wrong %v", any.AsString())
	assert(t, any.AsBlob() == nil, "wrong %v", any.AsBlob())
}
