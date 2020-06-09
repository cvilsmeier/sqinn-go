package sqinn

import (
	"testing"
)

func TestInt32(t *testing.T) {
	v := int(0x7FFFFFFF)
	buf := encodeInt32(v)
	assert(t, len(buf) == 4, "wrong %d", len(buf))
	assert(t, buf[0] == 0x7F, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0xFF, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0xFF, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0xFF, "wrong 0x%X", buf[3])
	var err error
	v, buf, err = decodeInt32(buf)
	assert(t, v == int(0x7FFFFFFF), "wrong %d", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
}

func TestInt64(t *testing.T) {
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
}

func TestDouble(t *testing.T) {
	v := float64(1.23)
	buf := encodeDouble(v)
	assert(t, len(buf) == 9, "wrong %d", len(buf))
	assert(t, buf[0] == 0x00, "wrong 0x%X", buf[0])
	assert(t, buf[1] == 0x00, "wrong 0x%X", buf[1])
	assert(t, buf[2] == 0x00, "wrong 0x%X", buf[2])
	assert(t, buf[3] == 0x05, "wrong 0x%X", buf[3])
	assert(t, buf[4] == 0x31, "wrong 0x%X", buf[4])
	assert(t, buf[5] == 0x2E, "wrong 0x%X", buf[5])
	assert(t, buf[6] == 0x32, "wrong 0x%X", buf[6])
	assert(t, buf[7] == 0x33, "wrong 0x%X", buf[7])
	assert(t, buf[8] == 0x00, "wrong 0x%X", buf[8])
	var err error
	v, buf, err = decodeDouble(buf)
	assert(t, v == float64(1.23), "wrong %g", v)
	assert(t, len(buf) == 0, "wrong %d", len(buf))
	assert(t, err == nil, "wrong %v", err)
}

func TestString(t *testing.T) {
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
}

func TestBlob(t *testing.T) {
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
}

func assert(t testing.TB, cond bool, format string, args ...interface{}) {
	t.Helper()
	if !cond {
		t.Fatalf(format, args...)
	}
}
