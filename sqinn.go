/*
Package sqinn provides interface to SQLite databases in Go without cgo.
It uses Sqinn (http://github.com/cvilsmeier/sqinn) for accessing SQLite
databases. It is not a database/sql driver.
*/
package sqinn

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/cvilsmeier/sqinn-go/v2/prebuilt"
)

// Options for launching a sqinn instance.
type Options struct {
	// Path to sqinn executable. Can be an absolute or relative path.
	// The name ":prebuilt:" is a special name that uses an embedded prebuilt
	// sqinn binary for linux/amd64 and windows/amd64.
	// Default is ":prebuilt:".
	Sqinn string

	// The database filename. It can be a file system path, e.g. "/tmp/test.db",
	// or a special name like ":memory:".
	// For further details, see https://www.sqlite.org/c3ref/open.html.
	// Default is ":memory:".
	Db string

	// The loglevel. Can be 0 (off), 1 (info) or 2 (debug).
	// Default is 0 (off).
	Loglevel int

	// Logfile is the filename to which sqinn will print log messages.
	// This is used for debugging and should normally be empty.
	// Default is empty (no log file).
	Logfile string

	// Log is a function that prints a log message from the sqinn process.
	// Log can be nil, then nothing will be logged
	// Default is nil (no logging).
	Log func(msg string)
}

// Prebuilt is a special path that tells sqinn-go to use an embedded
// pre-built sqinn binary. If Prebuilt is chosen, sqinn-go
// will extract sqinn into a temp directory and execute that.
// Not all os/arch combinations are embedded, though.
// Currently we have linux/amd64 and windows/amd64.
const Prebuilt string = ":prebuilt:"

// Sqinn is a running sqinn instance.
type Sqinn struct {
	tempdir string // only for prebuilt: tempdir where sqinn(.exe) is extracted
	cmd     *exec.Cmd
	mu      sync.Mutex
	w       *writer
	r       *reader
}

// Launch launches a new sqinn subprocess. The [Options] specify
// the sqinn executable, the database name, and logging options.
// See [Options] for details.
// If an error occurs, it returns (nil, err).
func Launch(opt Options) (*Sqinn, error) {
	if opt.Sqinn == "" {
		opt.Sqinn = Prebuilt
	}
	var tempdir string
	if opt.Sqinn == Prebuilt {
		dirname, filename, err := prebuilt.Extract()
		if err != nil {
			return nil, err
		}
		tempdir = dirname
		opt.Sqinn = filename
	}
	var cmdArgs []string
	cmdArgs = append(cmdArgs, "run")
	if opt.Db != "" {
		cmdArgs = append(cmdArgs, "-db", opt.Db)
	}
	if opt.Loglevel > 0 {
		cmdArgs = append(cmdArgs, "-loglevel", strconv.Itoa(opt.Loglevel))
	}
	if opt.Logfile != "" {
		cmdArgs = append(cmdArgs, "-logfile", opt.Logfile)
	}
	if opt.Log != nil {
		cmdArgs = append(cmdArgs, "-logstderr")
	}
	cmd := exec.Command(opt.Sqinn, cmdArgs...)
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
				opt.Log("[sqinn] " + sca.Text())
			}
			if err := sca.Err(); err != nil {
				opt.Log(fmt.Sprintf("cannot read sqinn stderr: %s", err))
			}
		}()
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &Sqinn{tempdir, cmd, sync.Mutex{}, writer, reader}, nil
}

// MustLaunch is the same as Launch except it panics on error.
func MustLaunch(opt Options) *Sqinn {
	return must(Launch(opt))
}

// ProduceFunc is a callback function that is called by Exec exactly
// once for each iteration.
// Iteration is the iteration index, starting at 0.
// Params is a slice that holds parameter values for this iteration,
// and must be set by the function body.
type ProduceFunc func(iteration int, params []Value)

// Exec executes a SQL statement, possibly multiple times.
//
// The niterations argument tells Exec how often to execute the SQL.
// It must be >= 0.
// If it is 0, Exec is a NO-OP.
//
// Binding SQL sql parameters is possible with the nparams argument
// and the produce function.
//
// The nparams argument tells Exec how many parameters to bind per iteration.
// It must be >= 0.
//
// The produce function produces parameter values. Parameter values can be
// of the following type: int, int64, float64, string, blob or nil.
// The length of the params argument is always nparams.
func (sq *Sqinn) Exec(sql string, niterations, nparams int, produce ProduceFunc) error {
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
		params := make([]Value, nparams)
		for iteration := range niterations {
			produce(iteration, params)
			sq.writeParams(params)
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

// MustExec is the same as Exec except it panics on error.
func (sq *Sqinn) MustExec(sql string, niterations, nparams int, produce ProduceFunc) {
	must(0, sq.Exec(sql, niterations, nparams, produce))
}

// ExecParams calls Exec with the provided params.
// The length of the params slice must be niterations * nparams.
func (sq *Sqinn) ExecParams(sql string, niterations, nparams int, params []Value) error {
	// check len(params)
	if len(params) != niterations*nparams {
		panic(fmt.Sprintf("want %d x %d params but have %d", niterations, nparams, len(params)))
	}
	// nothing to do if niterations is 0
	if niterations == 0 {
		return nil
	}
	return sq.Exec(sql, niterations, nparams, func(iteration int, iterationParams []Value) {
		if len(iterationParams) != nparams {
			panic(fmt.Sprintf("internal error: want %d iterationParams, but have only %d", nparams, len(iterationParams)))
		}
		offset := iteration * nparams
		n := copy(iterationParams, params[offset:offset+nparams])
		if n != nparams {
			panic(fmt.Sprintf("internal error: want %d params copied, but have only %d", nparams, n))
		}
	})
}

// MustExecParams is the same as ExecParams except it panics on error.
func (sq *Sqinn) MustExecParams(sql string, niterations, nparams int, params []Value) {
	must(0, sq.ExecParams(sql, niterations, nparams, params))
}

// ExecSql is the same as Exec(sql,1,0,nil).
func (sq *Sqinn) ExecSql(sql string) error {
	return sq.Exec(sql, 1, 0, nil)
}

// MustExecSql is the same as ExecSql except it panics on error.
func (sq *Sqinn) MustExecSql(sql string) {
	must(0, sq.ExecSql(sql))
}

// ConsumeFunc is a callback function that is called by Query once for each result row.
// Row is the row index, starting at 0.
// Values contains the row values for this row.
type ConsumeFunc func(row int, values []Value)

// Query executes a SQL statement and fetches the result rows.
//
// Params hold parameter values for the SQL statement. It can be empty.
//
// Coltypes defines the types of the columns to be fetched.
//
// Consume is called exactly once for each result row.
func (sq *Sqinn) Query(sql string, params []Value, coltypes []byte, consume ConsumeFunc) error {
	ncols := len(coltypes)
	if ncols == 0 {
		panic("no coltypes")
	}
	if consume == nil {
		panic("no consume func")
	}
	for _, param := range params {
		if param.Type == ValNull {
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

// MustQuery is the same as Query except it panics on error.
func (sq *Sqinn) MustQuery(sql string, params []Value, coltypes []byte, consume ConsumeFunc) {
	must(0, sq.Query(sql, params, coltypes, consume))
}

// QueryRows is like Query but consumes all rows and returns them in a [][]Value array.
func (sq *Sqinn) QueryRows(sql string, params []Value, coltypes []byte) ([][]Value, error) {
	var rows [][]Value
	err := sq.Query(sql, params, coltypes, func(row int, values []Value) {
		vals := append(make([]Value, 0, len(values)), values...)
		rows = append(rows, vals)
	})
	return rows, err
}

// MustQueryRows is the same as QueryRows except it panics on error.
func (sq *Sqinn) MustQueryRows(sql string, params []Value, coltypes []byte) [][]Value {
	return must(sq.QueryRows(sql, params, coltypes))
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
	if sq.tempdir != "" {
		os.RemoveAll(sq.tempdir)
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
	return fmt.Errorf("sqinn: %s", errmsg)
}

func (sq *Sqinn) writeParams(params []Value) {
	for _, p := range params {
		sq.w.writeByte(p.Type)
		switch p.Type {
		case ValNull:
			// no furhter data
		case ValInt32:
			sq.w.writeInt32(p.Int32)
		case ValInt64:
			sq.w.writeInt64(p.Int64)
		case ValDouble:
			sq.w.writeDouble(p.Double)
		case ValString:
			sq.w.writeString(p.String)
		case ValBlob:
			sq.w.writeBlob(p.Blob)
		default:
			panic(fmt.Sprintf("unknown param value type %T", p.Type))
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

// NullValue creates a Value with type ValNull.
func NullValue() Value { return Value{Type: ValNull} }

// Int32Value creates a Value with type ValInt32.
func Int32Value(v int) Value { return Value{Type: ValInt32, Int32: v} }

// Int64Value creates a Value with type ValInt64.
func Int64Value(v int64) Value { return Value{Type: ValInt64, Int64: v} }

// DoubleValue creates a Value with type ValDouble.
func DoubleValue(v float64) Value { return Value{Type: ValDouble, Double: v} }

// StringValue creates a Value with type ValString.
func StringValue(v string) Value { return Value{Type: ValString, String: v} }

// BlobValue creates a Value with type ValBlob.
func BlobValue(v []byte) Value { return Value{Type: ValBlob, Blob: v} }

// Value types.
const (
	ValNull   byte = 0
	ValInt32  byte = 1
	ValInt64  byte = 2
	ValDouble byte = 3
	ValString byte = 4
	ValBlob   byte = 5
)

// A Scanner scans Values.
type Scanner struct {
	values []Value
	i      int
}

// Scan creates a Scanner with the provided values.
func Scan(values []Value) *Scanner {
	return &Scanner{values, -1}
}

// Next returns the next Value.
func (s *Scanner) Next() Value {
	s.i++
	return s.values[s.i]
}

// NextInt32 returns the next values Int32 field and true if the value was not NULL, false if it was NULL.
func (s *Scanner) NextInt32() (int, bool) {
	v := s.Next()
	return v.Int32, v.Type != ValNull
}

// NextInt64 returns the next values Int64 field and true if the value was not NULL, false if it was NULL.
func (s *Scanner) NextInt64() (int64, bool) {
	v := s.Next()
	return v.Int64, v.Type != ValNull
}

// NextDouble returns the next values Double field and true if the value was not NULL, false if it was NULL.
func (s *Scanner) NextDouble() (float64, bool) {
	v := s.Next()
	return v.Double, v.Type != ValNull
}

// NextString returns the next values String field and true if the value was not NULL, false if it was NULL.
func (s *Scanner) NextString() (string, bool) {
	v := s.Next()
	return v.String, v.Type != ValNull
}

// NextBlob returns the next values Blob field and true if the value was not NULL, false if it was NULL.
func (s *Scanner) NextBlob() ([]byte, bool) {
	v := s.Next()
	return v.Blob, v.Type != ValNull
}

// Int32 returns the next values Int32 field. It does not check for NULL.
func (s *Scanner) Int32() int { return s.Next().Int32 }

// Int64 returns the next values Int64 field. It does not check for NULL.
func (s *Scanner) Int64() int64 { return s.Next().Int64 }

// Double returns the next values Double field. It does not check for NULL.
func (s *Scanner) Double() float64 { return s.Next().Double }

// String returns the next values String field. It does not check for NULL.
func (s *Scanner) String() string { return s.Next().String }

// Blob returns the next values Blob field. It does not check for NULL.
func (s *Scanner) Blob() []byte { return s.Next().Blob }

// Bind converts Go types to sqinn Values. It supports the following Go types
//
//	nil  -> ValNull
//	int  -> ValInt32
//	int64  -> ValInt64
//	float64  -> ValDouble
//	string  -> ValString
//	[]byte  -> ValBlob
//
// For any other Go type, it panics.
func Bind(params []any) []Value {
	if len(params) == 0 {
		return nil
	}
	values := make([]Value, len(params))
	for i, p := range params {
		if p == nil {
			values[i].Type = ValNull
			continue // with next param
		}
		switch v := p.(type) {
		case int:
			values[i].Type = ValInt32
			values[i].Int32 = v
		case int64:
			values[i].Type = ValInt64
			values[i].Int64 = v
		case float64:
			values[i].Type = ValDouble
			values[i].Double = v
		case string:
			values[i].Type = ValString
			values[i].String = v
		case []byte:
			values[i].Type = ValBlob
			values[i].Blob = v
		default:
			panic("sqinn.Bind(): wrong Go type")
		}
	}
	return values
}

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
