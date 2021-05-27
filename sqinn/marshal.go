package sqinn

import (
	"fmt"
	"math"
)

// byte marshalling

func decodeByte(buf []byte) (byte, []byte, error) {
	if len(buf) < 1 {
		return 0, nil, fmt.Errorf("cannot decodeByte from a %d byte buffer", len(buf))
	}
	v := buf[0]
	buf = buf[1:]
	return v, buf, nil
}

// bool marshalling

func decodeBool(buf []byte) (bool, []byte, error) {
	if len(buf) < 1 {
		return false, nil, fmt.Errorf("cannot decodeBool from a %d byte buffer", len(buf))
	}
	v := buf[0] != 0
	buf = buf[1:]
	return v, buf, nil
}

// int marshalling

func encodeInt32(v int) []byte {
	return []byte{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v >> 0),
	}
}

func decodeInt32(buf []byte) (int, []byte, error) {
	if len(buf) < 4 {
		return 0, nil, fmt.Errorf("cannot decodeInt32 from a %d byte buffer", len(buf))
	}
	v := int32(buf[0])<<24 |
		int32(buf[1])<<16 |
		int32(buf[2])<<8 |
		int32(buf[3])<<0
	buf = buf[4:]
	return int(v), buf, nil
}

// int64 marshalling

func encodeInt64(v int64) []byte {
	return []byte{
		byte(v >> 56),
		byte(v >> 48),
		byte(v >> 40),
		byte(v >> 32),
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v >> 0),
	}
}

func decodeInt64(buf []byte) (int64, []byte, error) {
	if len(buf) < 8 {
		return 0, nil, fmt.Errorf("cannot decodeInt64 from a %d byte buffer", len(buf))
	}
	v := int64(buf[0])<<56 |
		int64(buf[1])<<48 |
		int64(buf[2])<<40 |
		int64(buf[3])<<32 |
		int64(buf[4])<<24 |
		int64(buf[5])<<16 |
		int64(buf[6])<<8 |
		int64(buf[7])<<0
	buf = buf[8:]
	return v, buf, nil
}

// string marshalling

func encodeString(v string) []byte {
	data := []byte(v)
	sz := len(data) + 1
	buf := make([]byte, 0, 4+sz)
	buf = append(buf, encodeInt32(sz)...)
	buf = append(buf, data...)
	buf = append(buf, 0)
	return buf
}

func decodeString(buf []byte) (string, []byte, error) {
	sz, buf, err := decodeInt32(buf)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decodeString: %s", err)
	}
	if len(buf) < sz {
		return "", nil, fmt.Errorf("cannot decodeString length %d from a %d byte buffer", sz, len(buf))
	}
	v := string(buf[:sz-1])
	buf = buf[sz:]
	return v, buf, nil
}

// blob marshalling

func encodeBlob(v []byte) []byte {
	sz := len(v)
	buf := make([]byte, 0, 4+sz)
	buf = append(buf, encodeInt32(sz)...)
	buf = append(buf, v...)
	return buf
}

func decodeBlob(buf []byte) (_blob []byte, _buf []byte, _err error) {
	sz, buf, err := decodeInt32(buf)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decodeBlob: %s", err)
	}
	if len(buf) < sz {
		return nil, nil, fmt.Errorf("cannot decodeBlob length %d from a %d byte buffer", sz, len(buf))
	}
	v := buf[:sz]
	buf = buf[sz:]
	return v, buf, nil
}

// double marshalling

func encodeDouble(v float64) []byte {
	bv := math.Float64bits(v)
	return []byte{
		byte(bv >> 56),
		byte(bv >> 48),
		byte(bv >> 40),
		byte(bv >> 32),
		byte(bv >> 24),
		byte(bv >> 16),
		byte(bv >> 8),
		byte(bv >> 0),
	}
}

func decodeDouble(buf []byte) (float64, []byte, error) {
	if len(buf) < 8 {
		return 0, nil, fmt.Errorf("cannot decodeDouble from a %d byte buffer", len(buf))
	}
	bv := uint64(buf[0])<<56 |
		uint64(buf[1])<<48 |
		uint64(buf[2])<<40 |
		uint64(buf[3])<<32 |
		uint64(buf[4])<<24 |
		uint64(buf[5])<<16 |
		uint64(buf[6])<<8 |
		uint64(buf[7])<<0
	buf = buf[8:]
	v := math.Float64frombits(bv)
	return v, buf, nil
}
