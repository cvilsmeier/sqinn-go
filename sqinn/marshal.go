package sqinn

import (
	"fmt"
	"strconv"
)

// TODO cvvvvvvvv DO NOT PANIC!!

// byte marshalling

func decodeByte(buf []byte) (byte, []byte) {
	if len(buf) < 1 {
		panic(fmt.Errorf("cannot decodeByte from a %d byte buffer", len(buf)))
	}
	v := buf[0]
	buf = buf[1:]
	return v, buf
}

// bool marshalling

func decodeBool(buf []byte) (bool, []byte) {
	if len(buf) < 1 {
		panic(fmt.Errorf("cannot decodeBool from a %d byte buffer", len(buf)))
	}
	v := buf[0] != 0
	buf = buf[1:]
	return v, buf
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

func decodeInt32(buf []byte) (int, []byte) {
	if len(buf) < 4 {
		panic(fmt.Errorf("cannot decodeInt32 from a %d byte buffer", len(buf)))
	}
	v := int(buf[0])<<24 |
		int(buf[1])<<16 |
		int(buf[2])<<8 |
		int(buf[3])<<0
	buf = buf[4:]
	return v, buf
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

func decodeInt64(buf []byte) (int64, []byte) {
	if len(buf) < 8 {
		panic(fmt.Errorf("cannot decodeInt64 from a %d byte buffer", len(buf)))
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
	return v, buf
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

func decodeString(buf []byte) (string, []byte) {
	sz, buf := decodeInt32(buf)
	if len(buf) < sz {
		panic(fmt.Errorf("cannot decodeString length %d from a %d byte buffer", sz, len(buf)))
	}
	v := string(buf[:sz-1])
	buf = buf[sz:]
	return v, buf
}

// blob marshalling

func encodeBlob(v []byte) []byte {
	sz := len(v)
	buf := make([]byte, 0, 4+sz)
	buf = append(buf, encodeInt32(sz)...)
	buf = append(buf, v...)
	return buf
}

func decodeBlob(buf []byte) (_blob []byte, _buf []byte) {
	sz, buf := decodeInt32(buf)
	if len(buf) < sz {
		panic(fmt.Errorf("cannot decodeBlob length %d from a %d byte buffer", sz, len(buf)))
	}
	v := buf[:sz]
	buf = buf[sz:]
	return v, buf
}

// double marshalling

func encodeDouble(v float64) []byte {
	s := strconv.FormatFloat(v, 'g', -1, 64)
	return encodeString(s)
}

func decodeDouble(buf []byte) (float64, []byte) {
	s, buf := decodeString(buf)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(fmt.Errorf("cannot decodeDouble: %s", err))
	}
	return v, buf
}
