package sqinn

// value types, same as in sqinn/src/handler.h

type ValueType byte

// Value types for binding query parameters and retrieving column values.
const (

	// ValNull represents the NULL value (Go nil)
	ValNull ValueType = 0

	// ValInt represents a Go int
	ValInt ValueType = 1

	// ValInt64 represents a Go int64
	ValInt64 ValueType = 2

	// ValDouble represents a Go float64
	ValDouble ValueType = 6 // the IEEE variant

	// ValText represents a Go string
	ValText ValueType = 4

	// ValBlob represents a Go []byte
	ValBlob ValueType = 5
)

// An IntValue holds a nullable int value. The zero value is not set (a.k.a. NULL).
type IntValue struct {

	// Set is false if the value is NULL, otherwise true.
	Set bool

	// Value is the int value.
	Value int
}

// IsNull returns true if the value is NULL, otherwise true.
func (v IntValue) IsNull() bool { return !v.Set }

// An Int64Value holds a nullable int64 value. The zero value is not set (a.k.a. NULL).
type Int64Value struct {

	// Set is false if the value is NULL, otherwise true.
	Set bool

	// Value is the int64 value.
	Value int64
}

// IsNull returns true if the value is NULL, otherwise true.
func (v Int64Value) IsNull() bool { return !v.Set }

// A DoubleValue holds a nullable float64 value. The zero value is not set (a.k.a. NULL).
type DoubleValue struct {

	// Set is false if the value is NULL, otherwise true.
	Set bool

	// Value is the float64 value.
	Value float64
}

// IsNull returns true if the value is NULL, otherwise true.
func (v DoubleValue) IsNull() bool { return !v.Set }

// A StringValue holds a nullable string value. The zero value is not set (a.k.a. NULL).
type StringValue struct {

	// Set is false if the value is NULL, otherwise true.
	Set bool

	// Value is the string value.
	Value string
}

// IsNull returns true if the value is NULL, otherwise true.
func (v StringValue) IsNull() bool { return !v.Set }

// A BlobValue holds a nullable []byte value. The zero value is not set (a.k.a. NULL).
type BlobValue struct {
	// Set is false if the value is NULL, otherwise true.
	Set bool

	// Value is the []byte value.
	Value []byte
}

// IsNull returns true if the value is NULL, otherwise true.
func (v BlobValue) IsNull() bool { return !v.Set }

// An AnyValue can hold interface{} value type.
type AnyValue struct {
	Int    IntValue    // a nullable Go int
	Int64  Int64Value  // a nullable Go int64
	Double DoubleValue // a nullable Go float64
	String StringValue // a nullable Go string
	Blob   BlobValue   // a nullable Go []byte
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

func (a AnyValue) AsValue(t ValueType) interface{} {
	switch t {
	case ValInt:
		return a.Int.Value
	case ValInt64:
		return a.Int64.Value
	case ValDouble:
		return a.Double.Value
	case ValText:
		return a.String.Value
	case ValBlob:
		return a.Blob.Value
	}
	return nil
}

// A Row represents a query result row and holds a slice of values, one value
// per requested column.
type Row struct {
	Values []AnyValue
}
