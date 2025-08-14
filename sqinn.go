/*
Package sqinn provides interface to SQLite databases in Go without cgo.
It uses Sqinn2 (http://github.com/cvilsmeier/sqinn2) for accessing SQLite
databases. It is not a database/sql driver.
*/
package sqinn

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Options for launching a sqinn instance.
type Options struct {
	// Path to sqinn2 executable. Can be an absolute or relative path.
	// The name ":prebuilt:" is a special name that uses an embedded prebuilt
	// sqinn2 binary for linux/amd64 and windows/amd64.
	// Default is ":prebuilt:".
	Sqinn2 string

	// The loglevel. Can be 0 (off), 1 (info) or 2 (debug).
	// Default is 0 (off).
	Loglevel int

	// Logfile is the filename to which sqinn will print log messages.
	// This is used for debugging and should normally be empty.
	// Default no log file.
	Logfile string

	// Log is a function that prints a log message from the sqinn process.
	// Log can be nil, then nothing will be logged
	// Default is nil.
	Log func(msg string)

	// The database filename. It can be a file system path, e.g. "/tmp/test.db",
	// or a special name like ":memory:".
	// For further details, see https://www.sqlite.org/c3ref/open.html.
	// Default is ":memory:".
	Db string
}

// Prebuilt is a special path that tells sqinn-go to use an embedded
// pre-built sqinn2 binary. If Prebuilt is chosen, sqinn-go
// will extract sqinn2 into a temp directory and execute that.
// Not all os/arch combinations are embedded, though.
// Currently we have linux/amd64 and windows/amd64.
const Prebuilt = ":prebuilt:"

// Sqinn is a running sqinn instance.
type Sqinn struct {
	tempname string // for prebuilt
	cmd      *exec.Cmd
	mu       sync.Mutex
	w        *writer
	r        *reader
}

//go:embed "prebuilt/linux/sqinn2"
var prebuiltLinux []byte

//go:embed "prebuilt/windows/sqinn2.exe"
var prebuiltWindows []byte

// Launch launches a new sqinn2 subprocess. The [Options] specify
// the sqinn2 executable, the database name, and logging options.
// See [Options] for details.
// If an error occurs, it returns (nil, err).
func Launch(opt Options) (*Sqinn, error) {
	if opt.Sqinn2 == "" {
		opt.Sqinn2 = Prebuilt
	}
	var tempname string
	if opt.Sqinn2 == Prebuilt {
		prebuiltMap := map[string][]byte{
			"linux/amd64":   prebuiltLinux,
			"windows/amd64": prebuiltWindows,
		}
		filenameMap := map[string]string{
			"linux":   "sqinn2",
			"windows": "sqinn2.exe",
		}
		platform := runtime.GOOS + "/" + runtime.GOARCH
		prebuilt, prebuiltFound := prebuiltMap[platform]
		if !prebuiltFound {
			return nil, fmt.Errorf("no embedded prebuilt sqinn2 binary found for %s, please see https://github.com/cvilsmeier/sqinn2 for build instructions", platform)
		}
		tempdir, err := os.MkdirTemp("", "")
		if err != nil {
			return nil, err
		}
		tempname = filepath.Join(tempdir, filenameMap[runtime.GOOS])
		if err := os.WriteFile(tempname, prebuilt, 0755); err != nil {
			return nil, err
		}
		opt.Sqinn2 = tempname
	}
	var cmdArgs []string
	if opt.Loglevel > 0 {
		cmdArgs = append(cmdArgs, "-loglevel", strconv.Itoa(opt.Loglevel))
	}
	if opt.Logfile != "" {
		cmdArgs = append(cmdArgs, "-logfile", opt.Logfile)
	}
	if opt.Log != nil {
		cmdArgs = append(cmdArgs, "-logstderr")
	}
	if opt.Db != "" {
		cmdArgs = append(cmdArgs, "-db", opt.Db)
	}
	cmdArgs = append(cmdArgs, "-run")
	cmd := exec.Command(opt.Sqinn2, cmdArgs...)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	writer := newWriter(stdinPipe)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	reader := newReader(stdoutPipe)
	if opt.Log == nil {
		opt.Log = func(msg string) {}
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		opt.Log(fmt.Sprintf("cannot open sqinn stderr: %s", err))
	} else {
		go func() {
			sca := bufio.NewScanner(stderrPipe)
			for sca.Scan() {
				opt.Log("[sqinn2] " + sca.Text())
			}
			if err := sca.Err(); err != nil {
				opt.Log(fmt.Sprintf("cannot read sqinn stderr: %s", err))
			}
		}()
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &Sqinn{tempname, cmd, sync.Mutex{}, writer, reader}, nil
}

// MustLaunch is the same as Launch except it panics on error.
func MustLaunch(opt Options) *Sqinn {
	return must(Launch(opt))
}

// ExecRaw executes a SQL statement, possibly multiple times.
//
// The niterations argument tells Exec how often to execute the SQL.
// It must be >= 0.
// If it is 0, ExecRaw is a NO-OP.
//
// Binding SQL sql parameters is possible with the nparams argument and the produce function.
// The nparams argument tells ExecRaw how many parameters to bind per iteration.
// It must be >= 0.
//
// The produce function produces parameter values. Parameter values can be
// of the following type: int, int64, float64, string, blob or nil.
// The length of the params argument is always nparams.
func (sq *Sqinn) ExecRaw(sql string, niterations, nparams int, produce func(iteration int, params []any)) error {
	if niterations < 0 {
		panic("invalid niterations < 0")
	}
	if nparams < 0 {
		panic("invalid nparams < 0")
	}
	if nparams > 0 && produce == nil {
		panic("invalid nparams > 0 && produce == nil")
	}
	if niterations == 0 {
		return nil
	}
	sq.mu.Lock()
	defer sq.mu.Unlock()
	sq.w.writeByte(fcExec)       // FC_EXEC
	sq.w.writeString(sql)        // string sql
	sq.w.writeInt32(niterations) // int niterations
	sq.w.writeInt32(nparams)     // int nparams
	if nparams > 0 {
		params := make([]any, nparams)
		for iteration := range niterations {
			produce(iteration, params)
			sq.writeParams(params) // []value params
			if err := sq.w.markFrame(); err != nil {
				return err
			}
		}
	}
	if err := sq.w.flush(); err != nil {
		return err
	}
	return sq.readOk()
}

// Exec calls ExecRaw with the provided paramRows.
func (sq *Sqinn) Exec(sql string, paramRows [][]any) error {
	niterations := len(paramRows)
	if niterations == 0 {
		// nothing to do
		return nil
	}
	nparams := len(paramRows[0])
	// all paramRows must have same length
	for _, params := range paramRows {
		if len(params) != nparams {
			panic("all paramRows must have same length")
		}
	}
	return sq.ExecRaw(sql, niterations, nparams, func(iteration int, iterationParams []any) {
		n := copy(iterationParams, paramRows[iteration])
		if n != nparams {
			panic(fmt.Sprintf("internal error: want %d params copied, but have only %d", nparams, n))
		}
	})
}

// MustExec is the same as Exec except it panics on error.
func (sq *Sqinn) MustExec(sql string, paramRows [][]any) {
	must(0, sq.Exec(sql, paramRows))
}

// ExecSql is the same as Exec(sql,1,0,nil).
func (sq *Sqinn) ExecSql(sql string) error {
	return sq.ExecRaw(sql, 1, 0, nil)
}

// MustExecSql is the same as ExecSql except it panics on error.
func (sq *Sqinn) MustExecSql(sql string) {
	must(0, sq.ExecSql(sql))
}

// QueryRaw executes a SQL statement and fetches the result rows.
//
// Params hold parameter values fdr the SQL statement, if any.
//
// Coltypes defines the types of the columns to be fetched.
//
// Result row values are then fed into the consume function.
func (sq *Sqinn) QueryRaw(sql string, params []any, coltypes []byte, consume func(row int, values []Value)) error {
	ncols := len(coltypes)
	if ncols == 0 {
		panic("no coltypes")
	}
	if consume == nil {
		panic("no consume func")
	}
	for _, param := range params {
		if param == nil {
			panic("nil param not allowed in Query")
		}
	}
	for _, coltype := range coltypes {
		if coltype == ValNull {
			panic("coltype ValNull not allowed in Query")
		}
	}
	sq.mu.Lock()
	defer sq.mu.Unlock()
	sq.w.writeByte(fcQuery)        // FC_QUERY
	sq.w.writeString(sql)          // string sql
	sq.w.writeInt32(len(params))   // int nparams
	sq.writeParams(params)         // []value params
	sq.w.writeInt32(len(coltypes)) // int ncols
	for _, vt := range coltypes {  // []byte coltypes
		sq.w.writeByte(vt)
	}
	if err := sq.w.flush(); err != nil {
		return err
	}
	values := make([]Value, ncols)
	irow := -1
	for {
		irow++
		hasRow, err := sq.r.readByte()
		if err != nil {
			return err
		}
		if hasRow == 0 {
			break // no more rows
		}
		for icol := range coltypes {
			var val Value
			val.Type, err = sq.r.readByte()
			if err != nil {
				return err
			}
			switch val.Type {
			case ValNull:
				// not further data
			case ValInt32:
				val.Int32, err = sq.r.readInt32()
			case ValInt64:
				val.Int64, err = sq.r.readInt64()
			case ValDouble:
				val.Double, err = sq.r.readDouble()
			case ValString:
				val.String, err = sq.r.readString()
			case ValBlob:
				val.Blob, err = sq.r.readBlob()
			default:
				panic("invalid value type")
			}
			if err != nil {
				return err
			}
			values[icol] = val
		}
		consume(irow, values)
	}
	return sq.readOk()
}

// Query is like QueryRow but consumes all rows and returns them in a [][]Value array.
func (sq *Sqinn) Query(sql string, params []any, coltypes []byte) ([][]Value, error) {
	var rows [][]Value
	err := sq.QueryRaw(sql, params, coltypes, func(row int, values []Value) {
		vals := append(make([]Value, 0, len(values)), values...)
		rows = append(rows, vals)
	})
	return rows, err
}

// MustQuery is the same as Query except it panics on error.
func (sq *Sqinn) MustQuery(sql string, params []any, coltypes []byte) [][]Value {
	return must(sq.Query(sql, params, coltypes))
}

// Close closes the database and terminates the sqinn process.
func (sq *Sqinn) Close() error {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	sq.w.writeByte(fcQuit)
	if err := sq.w.flush(); err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	if err := sq.readOk(); err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	sq.cmd.WaitDelay = 5 * time.Second
	if err := sq.cmd.Wait(); err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	if sq.tempname != "" {
		os.Remove(sq.tempname)
	}
	return nil
}

func (sq *Sqinn) readOk() error {
	ok, err := sq.r.readByte()
	if err != nil {
		return err
	}
	if ok == 1 {
		return nil
	}
	errmsg, err := sq.r.readString()
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", errmsg)
}

func (sq *Sqinn) writeParams(params []any) {
	for _, p := range params {
		if p == nil {
			sq.w.writeByte(ValNull)
		} else {
			switch v := p.(type) {
			case int:
				sq.w.writeByte(ValInt32)
				sq.w.writeInt32(int(v))
			case int64:
				sq.w.writeByte(ValInt64)
				sq.w.writeInt64(v)
			case float64:
				sq.w.writeByte(ValDouble)
				sq.w.writeDouble(float64(v))
			case string:
				sq.w.writeByte(ValString)
				sq.w.writeString(v)
			case []byte:
				sq.w.writeByte(ValBlob)
				sq.w.writeBlob(v)
			default:
				panic(fmt.Sprintf("unknown param type %T", v))
			}
		}
	}
}

const (
	fcExec  = 1 // FC_EXEC
	fcQuery = 2 // FC_QUERY
	fcQuit  = 9 // FC_QUIT
)

// A Value holds a parameter or result value.
type Value struct {
	Type   byte    // The value type, can be any of ValNull, ValInt32, ValInt64, etc.
	Int32  int     // For ValInt32
	Int64  int64   // For ValInt64
	Double float64 // For ValDouble
	String string  // For ValString
	Blob   []byte  // For ValBlob
}

// Value types.
const (
	ValNull   = 0
	ValInt32  = 1
	ValInt64  = 2
	ValDouble = 3
	ValString = 4
	ValBlob   = 5
)

// A writer encodes values into bytes and writes them to a io.Writer.
type writer struct {
	w   io.Writer
	buf []byte
	wp  int // write pointer
}

func newWriter(w io.Writer) *writer {
	return &writer{w, make([]byte, 1024*1024), 0}
}

func (x *writer) append(p []byte) {
	n := len(p)
	minSize := x.wp + n
	for len(x.buf) < minSize {
		newSize := 2 * len(x.buf)
		for newSize < minSize {
			newSize = 2 * newSize
		}
		x.buf = append(x.buf, make([]byte, newSize-len(x.buf))...)
		// log.Printf("Writer.buf is now %d bytes", len(x.buf))
	}
	copy(x.buf[x.wp:], p)
	x.wp += n
}

func (x *writer) writeByte(v byte) {
	x.append([]byte{v})
}

func (x *writer) writeInt32(v int) {
	x.append(encodeInt32(v))
}

func (x *writer) writeInt64(v int64) {
	x.append(encodeInt64(v))
}

func (x *writer) writeDouble(v float64) {
	x.append(encodeDouble(v))
}

func (x *writer) writeString(v string) {
	x.writeInt32(len(v) + 1)
	x.append([]byte(v))
	x.append([]byte{0}) // null terminator
}

func (x *writer) writeBlob(v []byte) {
	x.writeInt32(len(v))
	x.append(v)
}

func (x *writer) markFrame() error {
	if x.wp > 1024*1024 {
		return x.flush()
	}
	return nil
}

func (x *writer) flush() error {
	if x.wp == 0 {
		return nil
	}
	lbuf := encodeInt32(x.wp)
	// log.Printf("to sqinn: %d len bytes: %v", len(lbuf), lbuf)
	n, err := x.w.Write(lbuf)
	if err != nil {
		return err
	}
	if n != 4 {
		return fmt.Errorf("want 4 byte write but was %d byte", n)
	}
	// log.Printf("to sqinn: %d bytes: %v", x.wp, x.buf[:x.wp])
	n, err = x.w.Write(x.buf[:x.wp])
	if err != nil {
		return err
	}
	if n != x.wp {
		return fmt.Errorf("want %d byte write but was %d byte", x.wp, n)
	}
	x.wp = 0
	return nil
}

// A reader reads bytes from a io.Reader and decodes them into values.
type reader struct {
	r    io.Reader
	buf  *bytes.Buffer
	buf1 []byte
	buf4 []byte
	buf8 []byte
}

func newReader(r io.Reader) *reader {
	return &reader{r, bytes.NewBuffer(nil), make([]byte, 1), make([]byte, 4), make([]byte, 8)}
}

func (x *reader) readBytes(n int) ([]byte, error) {
	avail := x.buf.Len()
	if avail == 0 {
		x.buf.Reset()
		if _, err := io.ReadFull(x.r, x.buf4); err != nil {
			return nil, err
		}
		n := decodeInt32(x.buf4)
		// log.Printf("read %d bytes from sqinn", n)
		if _, err := io.CopyN(x.buf, x.r, int64(n)); err != nil {
			return nil, err
		}
	}
	avail = x.buf.Len()
	if avail < n {
		return nil, fmt.Errorf("readBytes: want at least %d bytes available but have %d", n, avail)
	}
	var buf []byte
	switch n {
	case 1:
		buf = x.buf1
	case 4:
		buf = x.buf4
	case 8:
		buf = x.buf8
	default:
		buf = make([]byte, n)
	}
	if _, err := x.buf.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (x *reader) readByte() (byte, error) {
	buf, err := x.readBytes(1)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (x *reader) readInt32() (int, error) {
	buf, err := x.readBytes(4)
	if err != nil {
		return 0, err
	}
	return decodeInt32(buf), nil
}

func (x *reader) readInt64() (int64, error) {
	buf, err := x.readBytes(8)
	if err != nil {
		return 0, err
	}
	return decodeInt64(buf), nil
}

func (x *reader) readDouble() (float64, error) {
	buf, err := x.readBytes(8)
	if err != nil {
		return 0, err
	}
	return decodeDouble(buf), nil
}

func (x *reader) readString() (string, error) {
	buf, err := x.readBlob()
	if err != nil {
		return "", err
	}
	n := len(buf)
	if n < 1 {
		return "", fmt.Errorf("readString: invalid string length %d", n)
	}
	if buf[n-1] != 0 {
		return "", fmt.Errorf("readString: string must be null-terminated")
	}
	return string(buf[:n-1]), nil
}

func (x *reader) readBlob() ([]byte, error) {
	length, err := x.readInt32()
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return nil, nil
	}
	buf, err := x.readBytes(length)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// encode / decode

func encodeInt32(v int) []byte {
	return []byte{
		byte(uint32(v) >> 24),
		byte(uint32(v) >> 16),
		byte(uint32(v) >> 8),
		byte(uint32(v) >> 0),
	}
}

func decodeInt32(p []byte) int {
	i0 := int32(p[0]) << 24
	i1 := int32(p[1]) << 16
	i2 := int32(p[2]) << 8
	i3 := int32(p[3])
	return int(i0 + i1 + i2 + i3)
}

func encodeInt64(v int64) []byte {
	return []byte{
		byte(uint64(v) >> 56),
		byte(uint64(v) >> 48),
		byte(uint64(v) >> 40),
		byte(uint64(v) >> 32),
		byte(uint64(v) >> 24),
		byte(uint64(v) >> 16),
		byte(uint64(v) >> 8),
		byte(uint64(v) >> 0),
	}
}

func decodeInt64(p []byte) int64 {
	return (int64(p[0]))<<56 +
		(int64(p[1]))<<48 +
		(int64(p[2]))<<40 +
		(int64(p[3]))<<32 +
		(int64(p[4]))<<24 +
		(int64(p[5]))<<16 +
		(int64(p[6]))<<8 +
		(int64(p[7]))
}

func encodeDouble(v float64) []byte {
	b := math.Float64bits(v)
	return []byte{
		(byte)(b >> 56),
		(byte)(b >> 48),
		(byte)(b >> 40),
		(byte)(b >> 32),
		(byte)(b >> 24),
		(byte)(b >> 16),
		(byte)(b >> 8),
		(byte)(b >> 0),
	}
}

func decodeDouble(p []byte) float64 {
	i0 := (uint64(p[0])) << 56
	i1 := (uint64(p[1])) << 48
	i2 := (uint64(p[2])) << 40
	i3 := (uint64(p[3])) << 32
	i4 := (uint64(p[4])) << 24
	i5 := (uint64(p[5])) << 16
	i6 := (uint64(p[6])) << 8
	i7 := (uint64(p[7])) << 0
	return math.Float64frombits(i0 + i1 + i2 + i3 + i4 + i5 + i6 + i7)
}

// util

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}
