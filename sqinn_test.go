package sqinn

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestSqinn(t *testing.T) {
	opt := Options{
		Sqinn:    Prebuilt,
		Loglevel: 0,
		Log:      func(msg string) { t.Logf("SQINN: %s", msg) },
	}
	sq := MustLaunch(opt)
	t.Cleanup(func() {
		isNoErr(t, sq.Close())
	})
	//
	for _, sql := range []string{
		"PRAGMA journal_mode=DELETE",
		"PRAGMA synchronous=FULL",
		"PRAGMA foreign_keys=1",
		"PRAGMA busy_timeout=5000", // 5s
		"VACUUM",
		"CREATE TABLE users(i INTEGER, j INTEGER, d DOUBLE, t TEXT, b BLOB)",
	} {
		isNoErr(t, sq.ExecSql(sql))
	}
	//
	isNoErr(t, sq.Exec("INSERT INTO users (i,j,d,t,b) VALUES(?,?,?,?,?)", 3, 5, func(iteration int, params []Value) {
		isEq(t, 5, len(params))
		switch iteration {
		case 0:
			params[0] = Int32Value(1)
			params[1] = Int64Value(int64(1000))
			params[2] = DoubleValue(1.5)
			params[3] = StringValue("hi")
			params[4] = BlobValue([]byte("world1"))
		case 1:
			params[0] = Int32Value(2)
			params[1] = Int64Value(int64(2000))
			params[2] = DoubleValue(2.5)
			params[3] = StringValue("hi")
			params[4] = BlobValue([]byte("world2"))
		case 2:
			params[0] = NullValue()
			params[1] = NullValue()
			params[2] = NullValue()
			params[3] = NullValue()
			params[4] = NullValue()
		default:
			t.Fatal("wrong iteration ", iteration)
		}
	}))
	isNoErr(t, sq.Exec("UPDATE users SET t=? WHERE i=?", 2, 2, func(iteration int, params []Value) {
		isEq(t, 2, len(params))
		switch iteration {
		case 0:
			params[0] = StringValue("hello1")
			params[1] = Int32Value(1)
		case 1:
			params[0] = StringValue("hello2")
			params[1] = Int32Value(2)
		default:
			t.Fatal("wrong iteration ", iteration)
		}
	}))
	isNoErr(t, sq.ExecParams("UPDATE users SET t=? WHERE i=?", 2, 2, []Value{
		StringValue("hello1"), Int32Value(1),
		StringValue("hello2"), Int32Value(2),
	}))
	// Exec: niterations=0 is a NO-OP
	isNoErr(t, sq.Exec("UPDATE users SET t=? WHERE i=?", 0, 0, nil))
	isNoErr(t, sq.ExecParams("UPDATE users SET t=? WHERE i=?", 0, 1, []Value{}))
	//
	isNoErr(t, sq.Query("SELECT changes()", nil, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 1, values[0].Int32)
	}))
	//
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users", nil, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 3, values[0].Int32)
	}))
	//
	var rowCount int
	isNoErr(t, sq.Query("SELECT i,j,d,t,b FROM users ORDER BY i", nil, []byte{ValInt32, ValInt64, ValDouble, ValString, ValBlob}, func(row int, values []Value) {
		isEq(t, 5, len(values))
		switch row {
		case 0:
			isEq(t, ValNull, values[0].Type)
			isEq(t, ValNull, values[1].Type)
			isEq(t, ValNull, values[2].Type)
			isEq(t, ValNull, values[3].Type)
			isEq(t, ValNull, values[4].Type)
		case 1:
			isEq(t, 1, values[0].Int32)
			isEq(t, 1000, values[1].Int64)
			isEq(t, 1.5, values[2].Double)
			isEq(t, "hello1", values[3].String)
			isEq(t, "world1", string(values[4].Blob))
		case 2:
			isEq(t, 2, values[0].Int32)
			isEq(t, 2000, values[1].Int64)
			isEq(t, 2.5, values[2].Double)
			isEq(t, "hello2", values[3].String)
			isEq(t, "world2", string(values[4].Blob))
		default:
			t.Fatal("wrong row ", row)
		}
		rowCount++
	}))
	isEq(t, 3, rowCount)
	//
	// var rows [][]Value
	rows, err := sq.QueryRows("SELECT i,j,d,t,b FROM users ORDER BY i", nil, []byte{ValInt32, ValInt64, ValDouble, ValString, ValBlob})
	isNoErr(t, err)
	isEq(t, 3, len(rows))
	values := rows[0]
	isEq(t, 5, len(values))
	isEq(t, ValNull, values[0].Type)
	isEq(t, ValNull, values[1].Type)
	isEq(t, ValNull, values[2].Type)
	isEq(t, ValNull, values[3].Type)
	isEq(t, ValNull, values[4].Type)
	values = rows[1]
	isEq(t, 5, len(values))
	isEq(t, 1, values[0].Int32)
	isEq(t, 1000, values[1].Int64)
	isEq(t, 1.5, values[2].Double)
	isEq(t, "hello1", values[3].String)
	isEq(t, "world1", string(values[4].Blob))
	values = rows[2]
	isEq(t, 5, len(values))
	isEq(t, 2, values[0].Int32)
	isEq(t, 2000, values[1].Int64)
	isEq(t, 2.5, values[2].Double)
	isEq(t, "hello2", values[3].String)
	isEq(t, "world2", string(values[4].Blob))
	//
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users WHERE i = ?", []Value{Int32Value(2)}, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 1, values[0].Int32)
	}))
	//
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users WHERE j = ?", []Value{Int32Value(3)}, []byte{ValInt64}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 0, values[0].Int32)
	}))
	//
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users WHERE j = ?", []Value{Int32Value(2000)}, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 1, values[0].Int32)
	}))
	//
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users WHERE i IS NULL", nil, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 1, values[0].Int32)
	}))
	//
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users WHERE d = ?", []Value{DoubleValue(1.5)}, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 1, values[0].Int32)
	}))
	//
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users WHERE t = ?", []Value{StringValue("hello2")}, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 1, values[0].Int32)
	}))
	//
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users WHERE b = ?", []Value{BlobValue([]byte("world1"))}, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 1, values[0].Int32)
	}))
	//
	// to few params, but that's ok
	isNoErr(t, sq.Query("SELECT COUNT(*) FROM users WHERE i=?", nil, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, ValInt32, values[0].Type)
		isEq(t, 0, values[0].Int32)
	}))
	//
	// ExecRaw must panic if arguments are wrong
	isPanic(t, "invalid niterations < 0", func() {
		sq.Exec("DELETE FROM users WHERE i=131313", -1, -1, nil)
	})
	isPanic(t, "invalid nparams < 0", func() {
		sq.Exec("DELETE FROM users WHERE i=131313", 1, -1, nil)
	})
	isPanic(t, "invalid nparams > 0 && produce == nil", func() {
		sq.Exec("DELETE FROM users WHERE i=131313", 1, 1, nil)
	})
	//
	// Exec must panic if arguments are wrong
	isPanic(t, "want 1 x 1 params but have 2", func() {
		sq.ExecParams("DELETE FROM users WHERE i=?", 1, 1, []Value{Int32Value(0), NullValue()})
	})
	//
	// QueryRaw must panic if arguments are wrong
	isPanic(t, "coltype ValNull not allowed in Query", func() {
		sq.Query("SELECT COUNT(*) FROM users WHERE i = ?", nil, []byte{ValNull}, func(row int, values []Value) {})
	})
	isPanic(t, "nil param not allowed in Query", func() {
		sq.Query("SELECT COUNT(*) FROM users WHERE i=?", []Value{NullValue()}, []byte{ValInt32}, func(row int, values []Value) {})
	})
	isPanic(t, "no coltypes", func() {
		sq.Query("SELECT COUNT(*) FROM users", []Value{}, []byte{}, func(row int, values []Value) {})
	})
	isPanic(t, "no consume func", func() {
		sq.Query("SELECT COUNT(*) FROM users", []Value{}, []byte{ValInt32}, nil)
	})
	//
	// Query errors
	err = sq.Query("SELECT COUNT(*) FROM unknown_table_name WHERE hoob = 1", nil, []byte{ValInt32}, func(row int, values []Value) {})
	isErr(t, err, "sqinn: no such table: unknown_table_name")
	err = sq.Query("SELECT COUNT(*) FROM users WHERE hoob = 1", nil, []byte{ValInt32}, func(row int, values []Value) {})
	isErr(t, err, "sqinn: no such column: hoob")
	err = sq.Query("SELECT COUNT(*) FROM users", []Value{Int32Value(1)}, []byte{ValInt32, ValInt32}, func(row int, values []Value) {})
	isErr(t, err, "sqinn: column index out of range")
}

func TestSqinnLog(t *testing.T) {
	opt := Options{
		Sqinn:    "", // prebuilt
		Loglevel: 0,
		Log:      func(msg string) { t.Log("[SQINN] " + msg) },
		Db:       ":memory:",
	}
	sq := MustLaunch(opt)
	t.Cleanup(func() {
		err := sq.Close()
		isNoErr(t, err)
	})
	sq.MustExecSql("PRAGMA foreign_keys=1")
	err := sq.ExecParams("PRAGMA foreign_keys=1", 1, 0, []Value{})
	isNoErr(t, err)
	isNoErr(t, sq.Query("PRAGMA user_version", nil, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, 0, values[0].Int32)
	}))
}

func TestSqinnMust(t *testing.T) {
	opt := Options{}
	sq := MustLaunch(opt)
	t.Cleanup(func() {
		isNoErr(t, sq.Close())
	})
	sq.MustExec("PRAGMA foreign_keys=1", 1, 0, nil)
	sq.MustExecParams("PRAGMA foreign_keys=1", 1, 0, nil)
	sq.MustExecSql("PRAGMA foreign_keys=1")
	sq.MustQuery("PRAGMA user_version", nil, []byte{ValInt32}, func(row int, values []Value) {
		isEq(t, 0, row)
		isEq(t, 1, len(values))
		isEq(t, ValInt32, values[0].Type)
		isEq(t, 0, values[0].Int32)
	})
	rows := sq.MustQueryRows("PRAGMA user_version", nil, []byte{ValInt32})
	isEq(t, 1, len(rows))
	isEq(t, 1, len(rows[0]))
	isEq(t, ValInt32, rows[0][0].Type)
	isEq(t, 0, rows[0][0].Int32)
}

func TestSqinnBadPath(t *testing.T) {
	opt := Options{
		Sqinn:    "this_file_does_not_exist",
		Loglevel: 1,
		Logfile:  "/dev/null",
		Db:       ":memory:",
	}
	sq, err := Launch(opt)
	isEq(t, nil, sq)
	isTrue(t, err != nil, "want err but got nil")
	errmsg := err.Error()
	isTrue(t,
		errmsg == "exec: \"this_file_does_not_exist\": executable file not found in $PATH" ||
			errmsg == "exec: \"this_file_does_not_exist\": executable file not found in %PATH%",
		"invalid errmsg %q", errmsg)
}

func TestMemoryReaderWriter(t *testing.T) {
	byteValues := []byte{0, 127, 128, 255}
	int32Values := []int{0, 1, -1, 256, -256}
	int64Values := []int64{0, 1, -1, 256, -256}
	doubleValues := []float64{0, 128.5}
	stringValues := []string{"", "foobar"}
	blobValues := []string{"", "foobar", strings.Repeat("a", 1024*1024)}
	// write into memory
	wb := bytes.NewBuffer(nil)
	w := newWriter(wb)
	for _, value := range byteValues {
		w.writeByte(value)
	}
	for _, value := range int32Values {
		w.writeInt32(value)
	}
	for _, value := range int64Values {
		w.writeInt64(value)
	}
	for _, value := range doubleValues {
		w.writeDouble(value)
	}
	for _, value := range stringValues {
		w.writeString(value)
	}
	w.writeBlob(nil)
	for _, value := range blobValues {
		w.writeBlob([]byte(value))
	}
	w.writeString("end-marker")
	isNoErr(t, w.flush())
	// check memory content
	mem := wb.Bytes()
	isEq(t, 1048713, len(mem))
	var i int
	// 4 byte frame len
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x10, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x85, mem[i])
	i++
	isEq(t, 1048713-4, 0x100085)
	// byteValues := []byte{0, 127, 128, 255}
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x7F, mem[i])
	i++
	isEq(t, 0x80, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	// int32Values := []int{0, 1, -1, 256, -256}
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// 1
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x01, mem[i])
	i++
	// -1
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	// 256
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x01, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// -256
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// int64Values := []int64{0, 1, -1, 256, -256}
	// 0
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// 1
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x01, mem[i])
	i++
	// -1
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	// 256
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x01, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// -256}
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0xFF, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// doubleValues := []float64{0, 128.5}
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// double 128.5
	isEq(t, 0x40, mem[i])
	i++
	isEq(t, 0x60, mem[i])
	i++
	isEq(t, 0x10, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// string ""
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x01, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// string "foobar"
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x07, mem[i])
	i++
	isEq(t, 'f', mem[i])
	i++
	isEq(t, 'o', mem[i])
	i++
	isEq(t, 'o', mem[i])
	i++
	isEq(t, 'b', mem[i])
	i++
	isEq(t, 'a', mem[i])
	i++
	isEq(t, 'r', mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// blob nil
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// blob ""
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	// blob "foobar"
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x06, mem[i])
	i++
	isEq(t, 'f', mem[i])
	i++
	isEq(t, 'o', mem[i])
	i++
	isEq(t, 'o', mem[i])
	i++
	isEq(t, 'b', mem[i])
	i++
	isEq(t, 'a', mem[i])
	i++
	isEq(t, 'r', mem[i])
	i++
	// blob strings.Repeat("a", 1024*1024)}
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x10, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 'a', mem[i])
	i += 1024*1024 - 1 // and so on
	isEq(t, 'a', mem[i])
	i++
	// w.writeString("end-marker")
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, 0x0B, mem[i])
	i++
	isEq(t, 'e', mem[i])
	i++
	isEq(t, 'n', mem[i])
	i++
	isEq(t, 'd', mem[i])
	i++
	isEq(t, '-', mem[i])
	i++
	isEq(t, 'm', mem[i])
	i++
	isEq(t, 'a', mem[i])
	i++
	isEq(t, 'r', mem[i])
	i++
	isEq(t, 'k', mem[i])
	i++
	isEq(t, 'e', mem[i])
	i++
	isEq(t, 'r', mem[i])
	i++
	isEq(t, 0x00, mem[i])
	i++
	isEq(t, len(mem), i)
	// read from memory
	rb := bytes.NewBuffer(wb.Bytes())
	r := newReader(rb)
	for _, want := range byteValues {
		have, err := r.readByte()
		isNoErr(t, err)
		isEq(t, want, have)
	}
	for _, want := range int32Values {
		have, err := r.readInt32()
		isNoErr(t, err)
		isEq(t, want, have)
	}
	for _, want := range int64Values {
		have, err := r.readInt64()
		isNoErr(t, err)
		isEq(t, want, have)
	}
	for _, want := range doubleValues {
		have, err := r.readDouble()
		isNoErr(t, err)
		isEq(t, want, have)
	}
	for _, want := range stringValues {
		have, err := r.readString()
		isNoErr(t, err)
		isEq(t, want, have)
	}
	nilBlob, err := r.readBlob()
	isNoErr(t, err)
	isEq(t, 0, len(nilBlob))
	for _, want := range blobValues {
		have, err := r.readBlob()
		isNoErr(t, err)
		isEq(t, want, string(have))
	}
	str, err := r.readString()
	isNoErr(t, err)
	isEq(t, "end-marker", str)
}

func TestReadErrors(t *testing.T) {
	// read from memory
	rb := bytes.NewBuffer([]byte{
		0x00, 0x00, 0x00, 0x04, // frame length
		0x00, 0x00, 0x00, 0x00, // frame payload
	})
	r := newReader(rb)
	_, err := r.readInt64()
	isErr(t, err, "readBytes: want at least 8 bytes available but have 4")
	_, err = r.readString()
	isErr(t, err, "readString: invalid string length 0")
	// read from memory
	rb = bytes.NewBuffer([]byte{
		0x00, 0x00, // short frame length
	})
	r = newReader(rb)
	_, err = r.readInt64()
	isErr(t, err, "unexpected EOF")
	// read from memory
	rb = bytes.NewBuffer([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, // negative frame length
	})
	r = newReader(rb)
	_, err = r.readInt64()
	isErr(t, err, "readBytes: want at least 8 bytes available but have 0")
	// read from memory
	rb = bytes.NewBuffer([]byte{
		0x00, 0x00, 0x00, 0x07, // frame length
		0x00, 0x00, 0x00, 0x03, // string len
		0x41, 0x41, 0x41, // string data (no null term)
	})
	r = newReader(rb)
	_, err = r.readString()
	isErr(t, err, "readString: string must be null-terminated")
	_, err = r.readInt32()
	isErr(t, err, "EOF")
	_, err = r.readInt64()
	isErr(t, err, "EOF")
	_, err = r.readDouble()
	isErr(t, err, "EOF")
	_, err = r.readString()
	isErr(t, err, "EOF")
	_, err = r.readBlob()
	isErr(t, err, "EOF")
}

type errWriter struct {
	n   int
	err error
}

func (w *errWriter) Write(p []byte) (int, error) {
	if w.n == -1 {
		return len(p), w.err
	}
	return w.n, w.err
}

func TestWriterErrors(t *testing.T) {
	// write error
	w := newWriter(&errWriter{0, fmt.Errorf("fake error")})
	w.flush()
	w.writeInt32(1)
	w.writeString(strings.Repeat("a", 4*1024*1024))
	err := w.markFrame()
	isErr(t, err, "fake error")
	// short write
	w = newWriter(&errWriter{2, nil})
	err = w.flush()
	isNoErr(t, err)
	w.writeInt32(1)
	w.writeString(strings.Repeat("a", 4*1024*1024))
	err = w.markFrame()
	isErr(t, err, "want 4 byte write but was 2 byte")
	// short write
	ew := &errWriter{4, nil}
	w = newWriter(ew)
	err = w.flush()
	isNoErr(t, err)
	w.writeInt32(1)
	w.writeString(strings.Repeat("a", 4*1024*1024))
	err = w.markFrame()
	isErr(t, err, "want 4194313 byte write but was 4 byte")
	// short write
	ew = &errWriter{-1, nil}
	w = newWriter(ew)
	err = w.flush()
	isNoErr(t, err)
	w.writeInt32(1)
	w.writeString(strings.Repeat("a", 4*1024*1024))
	ew.err = fmt.Errorf("dummy write err")
	err = w.markFrame()
	isErr(t, err, "dummy write err")
}

func TestEncodeDecode(t *testing.T) {
	// int32
	p := encodeInt32(0)
	isEq(t, 0x00, p[0])
	isEq(t, 0x00, p[1])
	isEq(t, 0x00, p[2])
	isEq(t, 0x00, p[3])
	isEq(t, 0, decodeInt32(p))
	p = encodeInt32(1)
	isEq(t, 0x00, p[0])
	isEq(t, 0x00, p[1])
	isEq(t, 0x00, p[2])
	isEq(t, 0x01, p[3])
	isEq(t, 1, decodeInt32(p))
	p = encodeInt32(0x10203040)
	isEq(t, 0x10, p[0])
	isEq(t, 0x20, p[1])
	isEq(t, 0x30, p[2])
	isEq(t, 0x40, p[3])
	isEq(t, 270544960, decodeInt32(p))
	p = encodeInt32(-10)
	isEq(t, 0xFF, p[0])
	isEq(t, 0xFF, p[1])
	isEq(t, 0xFF, p[2])
	isEq(t, 0xF6, p[3])
	isEq(t, -10, decodeInt32(p))
	// int64
	p = encodeInt64(0)
	isEq(t, 0x00, p[4])
	isEq(t, 0x00, p[5])
	isEq(t, 0x00, p[6])
	isEq(t, 0x00, p[7])
	isEq(t, 0, decodeInt64(p))
	p = encodeInt64(1)
	isEq(t, 0x00, p[4])
	isEq(t, 0x00, p[5])
	isEq(t, 0x00, p[6])
	isEq(t, 0x01, p[7])
	isEq(t, 1, decodeInt64(p))
	p = encodeInt64(0x10203040)
	isEq(t, 0x10, p[4])
	isEq(t, 0x20, p[5])
	isEq(t, 0x30, p[6])
	isEq(t, 0x40, p[7])
	isEq(t, 270544960, decodeInt64(p))
	p = encodeInt64(-10)
	isEq(t, 0xFF, p[4])
	isEq(t, 0xFF, p[5])
	isEq(t, 0xFF, p[6])
	isEq(t, 0xF6, p[7])
	isEq(t, -10, decodeInt64(p))
	// double
	p = encodeDouble(0) // double 0.0 = hex(00 00 00 00 00 00 00 00)
	isEq(t, 0x00, p[0])
	isEq(t, 0x00, p[1])
	isEq(t, 0x00, p[2])
	isEq(t, 0x00, p[3])
	isEq(t, 0x00, p[4])
	isEq(t, 0x00, p[5])
	isEq(t, 0x00, p[6])
	isEq(t, 0x00, p[7])
	isEq(t, 0, decodeDouble(p))
	p = encodeDouble(128.5) // double 128.5 = hex(40 60 10 00 00 00 00 00)
	isEq(t, 0x40, p[0])
	isEq(t, 0x60, p[1])
	isEq(t, 0x10, p[2])
	isEq(t, 0x00, p[3])
	isEq(t, 0x00, p[4])
	isEq(t, 0x00, p[5])
	isEq(t, 0x00, p[6])
	isEq(t, 0x00, p[7])
	isEq(t, 128.5, decodeDouble(p))
	p = encodeDouble(-2.0) // double -2.0 = hex(C0 00 00 00 00 00 00 00)
	isEq(t, 0xC0, p[0])
	isEq(t, 0x00, p[1])
	isEq(t, 0x00, p[2])
	isEq(t, 0x00, p[3])
	isEq(t, 0x00, p[4])
	isEq(t, 0x00, p[5])
	isEq(t, 0x00, p[6])
	isEq(t, 0x00, p[7])
	isEq(t, -2.0, decodeDouble(p))
	p = encodeDouble(12345678.12345678) // double 12345678.12345678 = hex(41 67 8c 29 c3 f3 5b a2)
	isEq(t, 0x41, p[0])
	isEq(t, 0x67, p[1])
	isEq(t, 0x8C, p[2])
	isEq(t, 0x29, p[3])
	isEq(t, 0xC3, p[4])
	isEq(t, 0xF3, p[5])
	isEq(t, 0x5B, p[6])
	isEq(t, 0xA2, p[7])
	isEq(t, 12345678.12345678, decodeDouble(p))
	p = encodeDouble(-12345678.12345678) // double -12345678.12345678 = hex(c1 67 8c 29 c3 f3 5b a2)
	isEq(t, 0xC1, p[0])
	isEq(t, 0x67, p[1])
	isEq(t, 0x8C, p[2])
	isEq(t, 0x29, p[3])
	isEq(t, 0xC3, p[4])
	isEq(t, 0xF3, p[5])
	isEq(t, 0x5B, p[6])
	isEq(t, 0xA2, p[7])
	isEq(t, -12345678.12345678, decodeDouble(p))
}

func BenchmarkInsertUsers(b *testing.B) {
	const nusers = 10_000
	b.Run("Exec", func(b *testing.B) {
		dbfile := filepath.Join(b.TempDir(), "test.db")
		// launch sqinn
		sq := MustLaunch(Options{
			Db:       dbfile,
			Loglevel: 0,
			Log:      nil, // func(msg string) { b.Logf("SQINN: %s", msg) },
		})
		b.Cleanup(func() {
			err := sq.Close()
			if err != nil {
				b.Fatal(err)
			}
		})
		sq.MustExecSql("PRAGMA foreign_keys=1")
		sq.MustExecSql("CREATE TABLE users (id INTEGER NOT NULL PRIMARY KEY, name TEXT)")
		// benchmark: insert N users
		b.ResetTimer()
		for range b.N {
			sq.MustExecSql("DELETE FROM users")
			sq.MustExecSql("VACUUM")
			sq.MustExecSql("BEGIN IMMEDIATE")
			err := sq.Exec("INSERT INTO users(id,name) VALUES(?,?)", nusers, 2, func(iteration int, params []Value) {
				userId := iteration + 1
				params[0].Type = ValInt32
				params[0].Int32 = userId
				params[1].Type = ValString
				params[1].String = fmt.Sprintf("User %d", userId)
			})
			if err != nil {
				b.Fatal(err)
			}
			sq.MustExecSql("COMMIT")
		}
	})
	b.Run("ExecParams", func(b *testing.B) {
		dbfile := filepath.Join(b.TempDir(), "test.db")
		// launch sqinn
		sq := MustLaunch(Options{
			Db:       dbfile,
			Loglevel: 0,
			Log:      nil, // func(msg string) { b.Logf("SQINN: %s", msg) },
		})
		b.Cleanup(func() {
			err := sq.Close()
			if err != nil {
				b.Fatal(err)
			}
		})
		sq.MustExecSql("PRAGMA foreign_keys=1")
		sq.MustExecSql("CREATE TABLE users (id INTEGER NOT NULL PRIMARY KEY, name TEXT)")
		// benchmark: insert N users
		b.ResetTimer()
		for range b.N {
			sq.MustExecSql("DELETE FROM users")
			sq.MustExecSql("VACUUM")
			sq.MustExecSql("BEGIN IMMEDIATE")
			params := make([]Value, 0, nusers*2)
			for iuser := range nusers {
				userId := iuser + 1
				params = append(params, Int32Value(userId), StringValue(fmt.Sprintf("User %d", userId)))
			}
			sq.MustExecParams("INSERT INTO users(id,name) VALUES(?,?)", nusers, 2, params)
			sq.MustExecSql("COMMIT")
		}
	})
}

func BenchmarkQueryUsers(b *testing.B) {
	const nusers = 10_000
	b.Run("Query", func(b *testing.B) {
		dbfile := filepath.Join(b.TempDir(), "test.db")
		// launch sqinn
		sq := MustLaunch(Options{
			Db:       dbfile,
			Loglevel: 0,
			Log:      nil, // func(msg string) { b.Logf("SQINN: %s", msg) },
		})
		b.Cleanup(func() {
			err := sq.Close()
			if err != nil {
				b.Fatal(err)
			}
		})
		sq.MustExecSql("PRAGMA foreign_keys=1")
		sq.MustExecSql("CREATE TABLE users (id INTEGER NOT NULL PRIMARY KEY, name TEXT)")
		// insert N users
		sq.MustExecSql("BEGIN IMMEDIATE")
		var params []Value
		for iuser := range nusers {
			userId := iuser + 1
			params = append(params, Int32Value(userId), StringValue(fmt.Sprintf("User %d", userId)))
		}
		sq.MustExecParams("INSERT INTO users(id,name) VALUES(?,?)", nusers, 2, params)
		sq.MustExecSql("COMMIT")
		// benchmark: query N users
		b.ResetTimer()
		for range b.N {
			sq.MustExecSql("BEGIN IMMEDIATE")
			var rowCount int
			err := sq.Query("SELECT id,name FROM users ORDER BY id", nil, []byte{ValInt32, ValString}, func(row int, values []Value) {
				rowCount++
				if len(values) != 2 {
					b.Fatal("wrong len(values)")
				}
				id := values[0].Int32
				if id != row+1 {
					b.Fatal("wrong id")
				}
			})
			if err != nil {
				b.Fatal(err)
			}
			if rowCount != nusers {
				b.Fatal("wrong rowCount")
			}
			sq.MustExecSql("COMMIT")
		}
	})
	b.Run("MustQueryRows", func(b *testing.B) {
		dbfile := filepath.Join(b.TempDir(), "test.db")
		// launch sqinn
		sq := MustLaunch(Options{
			Db:       dbfile,
			Loglevel: 0,
			Log:      nil, // func(msg string) { b.Logf("SQINN: %s", msg) },
		})
		b.Cleanup(func() {
			err := sq.Close()
			if err != nil {
				b.Fatal(err)
			}
		})
		sq.MustExecSql("PRAGMA foreign_keys=1")
		sq.MustExecSql("CREATE TABLE users (id INTEGER NOT NULL PRIMARY KEY, name TEXT)")
		// insert N users
		sq.MustExecSql("BEGIN IMMEDIATE")
		var params []Value
		for iuser := range nusers {
			userId := iuser + 1
			params = append(params, Int32Value(userId), StringValue(fmt.Sprintf("User %d", userId)))
		}
		sq.MustExecParams("INSERT INTO users(id,name) VALUES(?,?)", nusers, 2, params)
		sq.MustExecSql("COMMIT")
		// benchmark: query N users
		b.ResetTimer()
		for range b.N {
			sq.MustExecSql("BEGIN IMMEDIATE")
			rows := sq.MustQueryRows("SELECT id,name FROM users ORDER BY id", nil, []byte{ValInt32, ValString})
			if len(rows) != nusers {
				b.Fatalf("wrong len(rows) %d", len(rows))
			}
			for irow, values := range rows {
				if len(values) != 2 {
					b.Fatal("wrong len(values)")
				}
				id := values[0].Int32
				if id != irow+1 {
					b.Fatal("wrong id")
				}
			}
			sq.MustExecSql("COMMIT")
		}
	})
}

// assertion library

func isTrue(t *testing.T, condition bool, format string, args ...any) {
	t.Helper()
	if !condition {
		t.Fatalf(format, args...)
	}
}

func isNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("want no err but have %s", err)
	}
}

func isErr(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("want err but have nil")
	} else if err.Error() != want {
		t.Fatalf("want err %q but have %q", want, err.Error())
	}
}

func isEq[T comparable](t *testing.T, want, have T) {
	t.Helper()
	if want != have {
		t.Fatalf("want %T(%v) but have %T(%v)", want, want, have, have)
	}
}

func isPanic(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("want panic but did not panic")
			return
		}
		isEq(t, want, r.(string))
	}()
	f()
	t.Fatalf("must not come here, want f() to panic")
}
