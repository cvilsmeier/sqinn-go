package sqinn

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestQueryRowsRetainsBlobs(t *testing.T) {
	sq := MustLaunch(Options{Db: filepath.Join(t.TempDir(), "blobs.db")})
	t.Cleanup(func() { isNoErr(t, sq.Close()) })
	sq.MustExecSql("CREATE TABLE blobs (id INTEGER PRIMARY KEY, b BLOB)")
	for _, size := range []int{1, 4, 8, 9} {
		first := bytes.Repeat([]byte{byte(0x40 + size)}, size)
		second := bytes.Repeat([]byte{byte(0x50 + size)}, size)
		sq.MustExecSql("DELETE FROM blobs")
		sq.MustExecParams("INSERT INTO blobs VALUES (?, ?)", 2, 2, []Value{
			Int32Value(1), BlobValue(first), Int32Value(2), BlobValue(second),
		})
		rows := sq.MustQueryRows("SELECT b, id FROM blobs ORDER BY id", nil, []byte{ValBlob, ValInt32})
		if !bytes.Equal(rows[0][0].Blob, first) || !bytes.Equal(rows[1][0].Blob, second) {
			t.Errorf("size %d: got blobs %x, %x; want %x, %x", size, rows[0][0].Blob, rows[1][0].Blob, first, second)
		}
		sq.MustQueryRows("SELECT 123456789", nil, []byte{ValInt64})
		if !bytes.Equal(rows[0][0].Blob, first) || !bytes.Equal(rows[1][0].Blob, second) {
			t.Errorf("size %d: later query changed retained blobs to %x, %x", size, rows[0][0].Blob, rows[1][0].Blob)
		}
	}
}
