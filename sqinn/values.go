package sqinn

// value types, same as in sqinn/src/handler.h

// Value types for binding query parameters and retrieving column values.
const (
	ValNull   byte = 0
	ValInt    byte = 1
	ValInt64  byte = 2
	ValDouble byte = 3
	ValText   byte = 4
	ValBlob   byte = 5
)

// An IntValue holds a nullable int value. The zero value is not set (a.k.a. NULL).
type IntValue struct {
	Set   bool
	Value int
}

// An Int64Value holds a nullable int64 value. The zero value is not set (a.k.a. NULL).
type Int64Value struct {
	Set   bool
	Value int64
}

// A DoubleValue holds a nullable float64 value. The zero value is not set (a.k.a. NULL).
type DoubleValue struct {
	Set   bool
	Value float64
}

// A StringValue holds a nullable string value. The zero value is not set (a.k.a. NULL).
type StringValue struct {
	Set   bool
	Value string
}

// A BlobValue holds a nullable []byte value. The zero value is not set (a.k.a. NULL).
type BlobValue struct {
	Set   bool
	Value []byte
}

// An AnyValue can hold any value type.
type AnyValue struct {
	Int    IntValue
	Int64  Int64Value
	Double DoubleValue
	String StringValue
	Blob   BlobValue
}

// AsInt returns an int value, or 0 if it is not set (NULL), or the value is not an int.
func (a AnyValue) AsInt() int {
	return a.Int.Value
}

// AsInt64 returns an int64 value or 0 if it is NULL or the value is not an int64.
func (a AnyValue) AsInt64() int64 {
	return a.Int64.Value
}

// AsDouble returns a double value or 0.0 if it is NULL or the value is not a double.
func (a AnyValue) AsDouble() float64 {
	return a.Double.Value
}

// AsString returns a string value or "" if it is NULL or the value is not a string.
func (a AnyValue) AsString() string {
	return a.String.Value
}

// AsBlob returns a []byte value or nil if it is NULL or the value is not a blob.
func (a AnyValue) AsBlob() []byte {
	return a.Blob.Value
}

// A Row holds many values.
type Row struct {
	Values []AnyValue
}
