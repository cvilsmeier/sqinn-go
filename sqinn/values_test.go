package sqinn_test

import (
	"testing"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func TestAnyValue(t *testing.T) {
	// IntValue represents a Go int
	any := sqinn.AnyValue{
		Int: sqinn.IntValue{
			Set:   true,
			Value: 13,
		},
	}
	assert(t, !any.Int.IsNull(), "wrong %v", any.Int.IsNull())
	assert(t, any.AsInt() == 13, "wrong %v", any.AsInt())
	assert(t, any.AsInt64() == 0, "wrong %v", any.AsInt64())
	assert(t, any.AsDouble() == 0, "wrong %v", any.AsDouble())
	assert(t, any.AsString() == "", "wrong %v", any.AsString())
	assert(t, any.AsBlob() == nil, "wrong %v", any.AsBlob())
	// Int64Value represents a Go int64
	any = sqinn.AnyValue{
		Int64: sqinn.Int64Value{
			Set:   true,
			Value: 42,
		},
	}
	assert(t, !any.Int64.IsNull(), "wrong %v", any.Int.IsNull())
	assert(t, any.AsInt() == 0, "wrong %v", any.AsInt())
	assert(t, any.AsInt64() == 42, "wrong %v", any.AsInt64())
	assert(t, any.AsDouble() == 0, "wrong %v", any.AsDouble())
	assert(t, any.AsString() == "", "wrong %v", any.AsString())
	assert(t, any.AsBlob() == nil, "wrong %v", any.AsBlob())
	// DoubleValue represents a Go float64
	any = sqinn.AnyValue{
		Double: sqinn.DoubleValue{
			Set:   true,
			Value: 42.42,
		},
	}
	assert(t, !any.Double.IsNull(), "wrong %v", any.Int.IsNull())
	assert(t, any.AsInt() == 0, "wrong %v", any.AsInt())
	assert(t, any.AsInt64() == 0, "wrong %v", any.AsInt64())
	assert(t, any.AsDouble() == 42.42, "wrong %v", any.AsDouble())
	assert(t, any.AsString() == "", "wrong %v", any.AsString())
	assert(t, any.AsBlob() == nil, "wrong %v", any.AsBlob())
	// StringValue represents a Go string
	any = sqinn.AnyValue{
		String: sqinn.StringValue{
			Set:   true,
			Value: "fourtytwo",
		},
	}
	assert(t, !any.String.IsNull(), "wrong %v", any.Int.IsNull())
	assert(t, any.AsInt() == 0, "wrong %v", any.AsInt())
	assert(t, any.AsInt64() == 0, "wrong %v", any.AsInt64())
	assert(t, any.AsDouble() == 0, "wrong %v", any.AsDouble())
	assert(t, any.AsString() == "fourtytwo", "wrong %v", any.AsString())
	assert(t, any.AsBlob() == nil, "wrong %v", any.AsBlob())
	// BlobValue represents a Go []byte
	any = sqinn.AnyValue{
		Blob: sqinn.BlobValue{
			Set:   true,
			Value: []byte{42, 43, 44},
		},
	}
	assert(t, !any.Blob.IsNull(), "wrong %v", any.Int.IsNull())
	assert(t, any.AsInt() == 0, "wrong %v", any.AsInt())
	assert(t, any.AsInt64() == 0, "wrong %v", any.AsInt64())
	assert(t, any.AsDouble() == 0, "wrong %v", any.AsDouble())
	assert(t, any.AsString() == "", "wrong %v", any.AsString())
	assert(t, any.AsBlob() != nil, "want != nil but was nil")
	assert(t, len(any.AsBlob()) == 3, "want len 3 but was %v", len(any.AsBlob()))
	assert(t, any.AsBlob()[0] == 42, "want [0] to be 42 but was %v", any.AsBlob()[0])
	assert(t, any.AsBlob()[1] == 43, "want [1] to be 43 but was %v", any.AsBlob()[1])
	assert(t, any.AsBlob()[2] == 44, "want [2] to be 44 but was %v", any.AsBlob()[2])
}
