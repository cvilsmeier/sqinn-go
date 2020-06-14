package sqinn

import (
	"testing"
)

func TestMarshalByte(t *testing.T) {
	var v byte
	buf := []byte{42}
	var err error
	v, buf, err = decodeByte(buf)
	assert(t, v == byte(42), "wrong %d", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeByte(buf)
	assert(t, v == byte(0), "wrong %d", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err != nil, "wrong %v", err)
}

func TestMarshalBool(t *testing.T) {
	var v bool
	buf := []byte{1}
	var err error
	v, buf, err = decodeBool(buf)
	assert(t, v == true, "wrong %v", v)
	assert(t, len(buf) == 0, "wrong %v", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeBool(buf)
	assert(t, v == false, "wrong %v", v)
	assert(t, len(buf) == 0, "wrong %v", len(buf))
	assert(t, err != nil, "wrong %v", err)
}

func TestMarshalInt32(t *testing.T) {
	v := int(0x7FFFFFFF)
	buf := encodeInt32(v)
	assert(t, len(buf) == 4, "wrong %v", len(buf))
	assert(t, buf[0] == 0x7F, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0xFF, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0xFF, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0xFF, "wrong 0x%X", buf[3])
	var err error
	v, buf, err = decodeInt32(buf)
	assert(t, v == int(0x7FFFFFFF), "wrong %d", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeInt32(buf)
	assert(t, v == int(0), "wrong %d", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err != nil, "wrong %v", err)
}

func TestMarshalInt64(t *testing.T) {
	v := int64(0x7FFFFFFFFFFFFF88)
	buf := encodeInt64(v)
	assert(t, len(buf) == 8, "wrong %d", len(buf))
	assert(t, buf[0] == 0x7F, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0xFF, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0xFF, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0xFF, "wrong 0x%X", buf[3])
	assert(t, buf[4] == 0xFF, "wrong 0x%X", buf[4])
	assert(t, buf[5] == 0xFF, "wrong 0x%X", buf[5])
	assert(t, buf[6] == 0xFF, "wrong 0x%X", buf[6])
	assert(t, buf[7] == 0x88, "wrong 0x%X", buf[7])
	var err error
	v, buf, err = decodeInt64(buf)
	assert(t, v == int64(0x7FFFFFFFFFFFFF88), "wrong %d", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeInt64(buf)
	assert(t, v == int64(0), "wrong %d", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err != nil, "wrong %v", err)
}

func TestMarshalDouble(t *testing.T) {
	// double 128.5 = hex(40 60 10 00 00 00 00 00)
	v := float64(128.5)
	buf := encodeDouble(v)
	assert(t, len(buf) == 8, "wrong %v", len(buf))
	assert(t, buf[0] == 0x40, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0x60, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0x10, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0x00, "wrong 0x%X", buf[3])
	assert(t, buf[4] == 0x00, "wrong 0x%X", buf[4])
	assert(t, buf[5] == 0x00, "wrong 0x%X", buf[5])
	assert(t, buf[6] == 0x00, "wrong 0x%X", buf[6])
	assert(t, buf[7] == 0x00, "wrong 0x%X", buf[7])
	var err error
	v, buf, err = decodeDouble(buf)
	assert(t, v == float64(128.5), "wrong %g", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeDouble(buf)
	assert(t, v == float64(0.0), "wrong %g", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err != nil, "wrong %v", err)
	// double 3.0 = hex(40 08 00 00 00 00 00 00)
	v = float64(3.0)
	buf = encodeDouble(v)
	assert(t, len(buf) == 8, "wrong %v", len(buf))
	assert(t, buf[0] == 0x40, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0x08, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0x00, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0x00, "wrong 0x%X", buf[3])
	assert(t, buf[4] == 0x00, "wrong 0x%X", buf[4])
	assert(t, buf[5] == 0x00, "wrong 0x%X", buf[5])
	assert(t, buf[6] == 0x00, "wrong 0x%X", buf[6])
	assert(t, buf[7] == 0x00, "wrong 0x%X", buf[7])
	v, buf, err = decodeDouble(buf)
	assert(t, v == float64(3.0), "wrong %g", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeDouble(buf)
	assert(t, v == float64(0.0), "wrong %g", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err != nil, "wrong %v", err)
	// double -2.0 = hex(C0 00 00 00 00 00 00 00)
	v = float64(-2.0)
	buf = encodeDouble(v)
	assert(t, len(buf) == 8, "wrong %v", len(buf))
	assert(t, buf[0] == 0xC0, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0x00, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0x00, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0x00, "wrong 0x%X", buf[3])
	assert(t, buf[4] == 0x00, "wrong 0x%X", buf[4])
	assert(t, buf[5] == 0x00, "wrong 0x%X", buf[5])
	assert(t, buf[6] == 0x00, "wrong 0x%X", buf[6])
	assert(t, buf[7] == 0x00, "wrong 0x%X", buf[7])
	v, buf, err = decodeDouble(buf)
	assert(t, v == float64(-2.0), "wrong %g", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeDouble(buf)
	assert(t, v == float64(0.0), "wrong %g", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err != nil, "wrong %v", err)
}

func TestMarshalString(t *testing.T) {
	v := "foo"
	buf := encodeString(v)
	assert(t, len(buf) == 8, "wrong %d", len(buf))
	assert(t, buf[0] == 0x00, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0x00, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0x00, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0x04, "wrong 0x%X", buf[3])
	assert(t, buf[4] == 0x66, "wrong 0x%X", buf[4])
	assert(t, buf[5] == 0x6F, "wrong 0x%X", buf[5])
	assert(t, buf[6] == 0x6F, "wrong 0x%X", buf[6])
	assert(t, buf[7] == 0x00, "wrong 0x%X", buf[7])
	var err error
	v, buf, err = decodeString(buf)
	assert(t, v == "foo", "wrong %q", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeString(buf)
	assert(t, len(v) == 0, "wrong %d", len(v))
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err != nil, "wrong %v", err)
}

func TestMarshalBlob(t *testing.T) {
	v := []byte{1, 2, 3}
	buf := encodeBlob(v)
	assert(t, len(buf) == 7, "wrong %d", len(buf))
	assert(t, buf[0] == 0x00, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0x00, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0x00, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0x03, "wrong 0x%X", buf[3])
	assert(t, buf[4] == 0x01, "wrong 0x%X", buf[4])
	assert(t, buf[5] == 0x02, "wrong 0x%X", buf[5])
	assert(t, buf[6] == 0x03, "wrong 0x%X", buf[6])
	var err error
	v, buf, err = decodeBlob(buf)
	assert(t, len(v) == 3, "wrong %d", len(v))
	assert(t, v[0] == 1, "wrong %d", v[0])
	assert(t, v[1] == 2, "wrong %d", v[1])
	assert(t, v[2] == 3, "wrong %d", v[2])
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
	v, buf, err = decodeBlob(buf)
	assert(t, len(v) == 0, "wrong %d", len(v))
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err != nil, "wrong %v", err)
}

func assert(t testing.TB, cond bool, format string, args ...interface{}) {
	t.Helper()
	if !cond {
		t.Fatalf(format, args...)
	}
}
