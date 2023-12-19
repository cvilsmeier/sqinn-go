package sqinn

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"unsafe"
)

// function codes, same as in sqinn/src/handler.h

const (
	fcSqinnVersion  byte = 1
	fcIoVersion     byte = 2
	fcSqliteVersion byte = 3
	fcOpen          byte = 10
	fcPrepare       byte = 11
	fcBind          byte = 12
	fcStep          byte = 13
	fcReset         byte = 14
	fcChanges       byte = 15
	fcColumn        byte = 16
	fcFinalize      byte = 17
	fcClose         byte = 18
	fcColumnCount   byte = 19
	fcColumnType    byte = 20
	fcColumnName    byte = 21
	fcExec          byte = 51
	fcQuery         byte = 52
)

// Options for launching a Sqinn instance.
type Options struct {

	// Path to Sqinn executable. Can be an absolute or relative path.
	// Empty is the same as "sqinn". Default is empty.
	SqinnPath string

	// Logger logs the debug and error messages that the sinn subprocess will output
	// on its stderr. Default is nil, which does not log anything.
	Logger Logger

	// Log the binary io protocol. Only for debugging. Should normally be false. Default is false.
	LogBinary bool
}

// Sqinn is a running sqinn instance.
type Sqinn struct {
	mx        sync.Mutex
	logBinary bool
	cmd       *exec.Cmd
	sin       io.WriteCloser
	sout      io.ReadCloser
	serr      io.ReadCloser
}

// Launch launches a new Sqinn subprocess. The options specify
// the path to the sqinn executable, among others. See docs for
// Options for details.
// If an error occurs, it returns (nil, err).
func Launch(options Options) (*Sqinn, error) {
	sqinnPath := options.SqinnPath
	if sqinnPath == "" {
		sqinnPath = "sqinn"
	}
	cmd := exec.Command(sqinnPath)
	sin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	sout, err := cmd.StdoutPipe()
	if err != nil {
		sin.Close()
		return nil, err
	}
	serr, err := cmd.StderrPipe()
	if err != nil {
		sout.Close()
		sin.Close()
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		serr.Close()
		sout.Close()
		sin.Close()
		return nil, err
	}
	sq := &Sqinn{sync.Mutex{}, options.LogBinary, cmd, sin, sout, serr}
	logger := options.Logger
	if logger == nil {
		logger = NoLogger{}
	}
	go sq.run(logger)
	return sq, nil
}

// MustLaunch is like Launch except it panics on error.
func MustLaunch(options Options) *Sqinn {
	sq, err := Launch(options)
	if err != nil {
		panic(err)
	}
	return sq
}

func (sq *Sqinn) run(logger Logger) {
	sc := bufio.NewScanner(sq.serr)
	for sc.Scan() {
		text := sc.Text()
		logger.Log(fmt.Sprintf("[sqinn] %s", text))
	}
	err := sc.Err()
	if err != nil {
		logger.Log(fmt.Sprintf("[sqinn] stderr: %s", err))
	}
}

// SqinnVersion returns the version of the Sqinn executable.
// If an error occurs, it returns ("", err).
func (sq *Sqinn) SqinnVersion() (string, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcSqinnVersion}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return "", err
	}
	var version string
	version, _, err = decodeString(resp)
	if err != nil {
		return "", err
	}
	return version, nil
}

// IoVersion returns the protocol version for this Sqinn instance.
// The version is >= 1.
// If an error occurs, it returns (0, err).
func (sq *Sqinn) IoVersion() (byte, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcIoVersion}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return 0, err
	}
	var version byte
	version, _, err = decodeByte(resp)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// SqliteVersion returns the SQLite library version Sqinn was built with.
// If an error occurs, it returns ("", err).
func (sq *Sqinn) SqliteVersion() (string, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcSqliteVersion}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return "", err
	}
	var version string
	version, _, err = decodeString(resp)
	if err != nil {
		return "", err
	}
	return version, nil
}

// Open opens a database.
// The filename can be ":memory:" or interface{} filesystem path, e.g. "/tmp/test.db".
// Sqinn keeps the database open until Close is called. After Close has been
// called, this Sqinn instance can be terminated with Terminate, or Open can be
// called again, either on the same database or on a different one. For every
// Open there should be a Close call.
//
// For further details, see https://www.sqlite.org/c3ref/open.html.
func (sq *Sqinn) Open(filename string) error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 10+len(filename))
	req = append(req, fcOpen)
	req = append(req, encodeString(filename)...)
	// resp
	_, err := sq.writeAndRead(req)
	if err != nil {
		return err
	}
	return nil
}

// MustOpen is like Open except it panics on error.
func (sq *Sqinn) MustOpen(filename string) {
	err := sq.Open(filename)
	if err != nil {
		panic(err)
	}
}

// Prepare prepares a statement, using the provided sql string.
// To avoid memory leaks, each prepared statement must be finalized
// after use. Sqinn allows only one prepared statement at at time,
// preparing a statement while another statement is still active
// (not yet finalized) will result in a error.
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/prepare.html.
func (sq *Sqinn) Prepare(sql string) error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 10+len(sql))
	req = append(req, fcPrepare)
	req = append(req, encodeString(sql)...)
	// resp
	_, err := sq.writeAndRead(req)
	if err != nil {
		return err
	}
	return nil
}

func (sq *Sqinn) bindValue(req []byte, value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		req = append(req, byte(ValNull))
	case int:
		req = append(req, byte(ValInt))
		req = append(req, encodeInt32(v)...)
	case int64:
		req = append(req, byte(ValInt64))
		req = append(req, encodeInt64(v)...)
	case float64:
		req = append(req, byte(ValDouble))
		req = append(req, encodeDouble(float64(v))...)
	case string:
		req = append(req, byte(ValText))
		req = append(req, encodeString(v)...)
	case []byte:
		req = append(req, byte(ValBlob))
		req = append(req, encodeBlob(v)...)
	default:
		return nil, fmt.Errorf("cannot bind type %T", v)
	}
	return req, nil
}

func (sq *Sqinn) bindValues(req []byte, values []interface{}) ([]byte, error) {
	var err error
	for _, value := range values {
		req, err = sq.bindValue(req, value)
		if err != nil {
			return nil, err
		}
	}
	return req, nil
}

// Bind binds the iparam'th parameter with the specified value.
// The value can be an int, int64, float64, string, []byte or nil.
// Not that iparam starts at 1 (not 0):
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/bind_blob.html.
func (sq *Sqinn) Bind(iparam int, value interface{}) error {
	if iparam < 1 {
		return fmt.Errorf("Bind: iparam must be >= 1 but was %d", iparam)
	}
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0)
	req = append(req, fcBind)
	req = append(req, encodeInt32(iparam)...)
	var err error
	req, err = sq.bindValue(req, value)
	if err != nil {
		return err
	}
	// resp
	_, err = sq.writeAndRead(req)
	if err != nil {
		return err
	}
	return nil
}

// Step advances the current statement to the next row or to completion.
// It returns true if there are more rows available, false if not.
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/step.html.
func (sq *Sqinn) Step() (bool, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcStep}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return false, err
	}
	more, _, err := decodeBool(resp)
	if err != nil {
		return false, err
	}
	return more, nil
}

// Reset resets the current statement to its initial state.
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/reset.html.
func (sq *Sqinn) Reset() error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcReset}
	// resp
	_, err := sq.writeAndRead(req)
	if err != nil {
		return err
	}
	return nil
}

// Changes counts the number of rows modified by the last SQL operation.
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/changes.html.
func (sq *Sqinn) Changes() (int, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcChanges}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return 0, err
	}
	var changes int
	changes, _, err = decodeInt32(resp)
	if err != nil {
		return 0, err
	}
	return changes, nil
}

// ColumnsCount returns the number of columns in the result set.
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/column_count.html.
func (sq *Sqinn) ColumnCount() (int, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcColumnCount}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return 0, err
	}
	var columnsCount int
	columnsCount, _, err = decodeInt32(resp)
	if err != nil {
		return 0, err
	}
	return columnsCount, nil
}

// ColumnType returns the type of the specified column in the result set.
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/column_blob.html.
func (sq *Sqinn) ColumnType(col int) (ValueType, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 4)
	req = append(req, fcColumnType)
	req = append(req, encodeInt32(col)...)
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return ValNull, err
	}
	colType, _, err := decodeByte(resp)
	if err != nil {
		return ValNull, err
	}
	return ValueType(colType), nil
}

func (sq *Sqinn) ColumnName(col int) (string, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 2)
	req = append(req, fcColumnName)
	req = append(req, encodeInt32(col)...)
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return "", err
	}
	colName, _, err := decodeString(resp)
	if err != nil {
		return "", err
	}
	return colName, nil
}

func (sq *Sqinn) decodeAnyValue(resp []byte, colType ValueType) (AnyValue, []byte, error) {
	var any AnyValue
	var set bool
	var err error
	set, resp, err = decodeBool(resp)
	if err != nil {
		return any, nil, err
	}
	if set {
		switch colType {
		case ValNull:
			err = fmt.Errorf("ValNull is not a valid column type")
		case ValInt:
			any.Int.Set = true
			any.Int.Value, resp, err = decodeInt32(resp)
		case ValInt64:
			any.Int64.Set = true
			any.Int64.Value, resp, err = decodeInt64(resp)
		case ValDouble:
			any.Double.Set = true
			any.Double.Value, resp, err = decodeDouble(resp)
		case ValText:
			any.String.Set = true
			any.String.Value, resp, err = decodeString(resp)
		case ValBlob:
			any.Blob.Set = true
			any.Blob.Value, resp, err = decodeBlob(resp)
		default:
			err = fmt.Errorf("invalid col type %d", colType)
		}
		if err != nil {
			return any, nil, err
		}
	}
	return any, resp, nil
}

// Column retrieves the value of the icol'th column.
// The colType specifies the expected type of the column value.
// Note that icol starts at 0 (not 1).
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/column_blob.html.
func (sq *Sqinn) Column(icol int, colType ValueType) (AnyValue, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 6)
	req = append(req, fcColumn)
	req = append(req, encodeInt32(icol)...)
	req = append(req, byte(colType))
	// resp
	var any AnyValue
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return any, err
	}
	any, _, err = sq.decodeAnyValue(resp, colType)
	return any, err
}

// Finalize finalizes a statement that has been prepared with Prepare.
// To avoid memory leaks, each statement has to be finalized.
// Moreover, since Sqinn allows only one statement at a time,
// each statement must be finalized before a new statement can be prepared.
//
// This is a low-level function. Use Exec/Query instead.
//
// For further details, see https://www.sqlite.org/c3ref/finalize.html.
func (sq *Sqinn) Finalize() error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcFinalize}
	// resp
	_, err := sq.writeAndRead(req)
	return err
}

// Close closes the database connection that has been opened with Open.
// After Close has been called, this Sqinn instance can be terminated, or
// another database can be opened with Open.
//
// For further details, see https://www.sqlite.org/c3ref/close.html.
func (sq *Sqinn) Close() error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{fcClose}
	// resp
	_, err := sq.writeAndRead(req)
	return err
}

// ExecOne executes a SQL statement and returns the number of modified rows.
// It is used primarily for short, simple statements that have no parameters
// and do not query rows. A good use case is for beginning and committing
// a transaction:
//
//	_, err = sq.ExecOne("BEGIN");
//	// do stuff in tx
//	_, err = sq.ExecOne("COMMIT");
//
// Another use case is for DDL statements:
//
//	_, err = sq.ExecOne("DROP TABLE users");
//	_, err = sq.ExecOne("CREATE TABLE foo (name VARCHAR)");
//
// ExecOne(sql) has the same effect as Exec(sql, 1, 0, nil).
//
// If a error occurs, ExecOne will return (0, err).
func (sq *Sqinn) ExecOne(sql string) (int, error) {
	changes, err := sq.Exec(sql, 1, 0, nil)
	if err != nil {
		return 0, err
	}
	return changes[0], nil
}

// MustExecOne is like ExecOne except it panics on error.
func (sq *Sqinn) MustExecOne(sql string) int {
	mod, err := sq.ExecOne(sql)
	if err != nil {
		panic(err)
	}
	return mod
}

// Exec executes a SQL statement multiple times and returns the
// number of modified rows for each iteration. It supports bind parmeters.
// Exec is used to execute SQL statements that do not return results (see
// Query for those).
//
// The niterations tells Exec how often to run the sql. It must be >= 0 and
// should be >= 1. If niterations is zero, the statement is not run at all,
// and the method call is a waste of CPU cycles.
//
// Binding sql parameters is possible with the nparams and values arguments.
// The nparams argument tells Exec how many parameters to bind per iteration.
// nparams must be >= 0.
//
// The values argument holds the parameter values. Parameter values can be
// of the following type: int, int64, float64, string, blob or nil.
// The length of values must always be niterations * nparams.
//
// Internally, Exec preapres a statement, binds nparams parameters, steps
// the statement, resets the statement, binds the next nparams parameters,
// and so on, until niterations is reached.
//
// Exec returns, for each iteration, the count of modified rows. The
// resulting int slice will always be of length niterations.
//
// If an error occurs, it will return (nil, err).
func (sq *Sqinn) Exec(sql string, niterations, nparams int, values []interface{}) ([]int, error) {
	if niterations < 0 {
		return nil, fmt.Errorf("Exec '%s' niterations must be >= 0 but was %d", sql, niterations)
	}
	if len(values) != niterations*nparams {
		return nil, fmt.Errorf("Exec '%s' expected %d values but have %d", sql, niterations*nparams, len(values))
	}
	sq.mx.Lock()
	defer sq.mx.Unlock()
	req := make([]byte, 0, len(sql)+10*len(values))
	req = append(req, fcExec)
	req = append(req, encodeString(sql)...)
	req = append(req, encodeInt32(niterations)...)
	req = append(req, encodeInt32(nparams)...)
	var err error
	req, err = sq.bindValues(req, values)
	if err != nil {
		return nil, err
	}
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return nil, err
	}
	changes := make([]int, niterations)
	for i := 0; i < niterations; i++ {
		changes[i], resp, err = decodeInt32(resp)
		if err != nil {
			return nil, err
		}
	}
	return changes, nil
}

// MustExec is like Exec except it panics on error.
func (sq *Sqinn) MustExec(sql string, niterations, nparams int, values []interface{}) []int {
	mods, err := sq.Exec(sql, niterations, nparams, values)
	if err != nil {
		panic(err)
	}
	return mods
}

// Query executes a SQL statement and returns all rows.
// Query is used for SELECT statements.
//
// The params argument holds a list of bind parameters. Values must be of type
// int, int64, float64, string, []byte or nil.
//
// The colTypes argument holds a list of column types that the query yields.
//
// Query returns all resulting rows at once. There is no way
// to interrupt a Query while it is running. If a Query yields more data
// than can fit into memory, the behavior is undefined, most likely an
// out-of-memory condition will crash your program. It is up to the caller to
// make sure that all queried data fits into memory. The sql 'LIMIT' operator
// may be helpful.
//
// Each returned Row contains a slice of values. The number of values per row is
// equal to the length of colTypes.
//
// If an error occurs, it will return (nil, err).
func (sq *Sqinn) Query(sql string, params []interface{}, colTypes []ValueType) ([]Row, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	req := make([]byte, 0, len(sql)+8*len(params))
	req = append(req, fcQuery)
	req = append(req, encodeString(sql)...)
	nparams := len(params)
	req = append(req, encodeInt32(nparams)...)
	var err error
	req, err = sq.bindValues(req, params)
	if err != nil {
		return nil, err
	}
	ncols := len(colTypes)
	byteCols := *(*[]byte)(unsafe.Pointer(&colTypes))
	req = append(req, encodeInt32(ncols)...)
	req = append(req, byteCols...)
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return nil, err
	}
	var nrows int
	nrows, resp, err = decodeInt32(resp)
	if err != nil {
		return nil, err
	}
	rows := make([]Row, 0, nrows)
	for i := 0; i < nrows; i++ {
		var row Row
		row.Values = make([]AnyValue, 0, ncols)
		for icol := 0; icol < ncols; icol++ {
			var any AnyValue
			any, resp, err = sq.decodeAnyValue(resp, colTypes[icol])
			if err != nil {
				return nil, err
			}
			row.Values = append(row.Values, any)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// MustQuery is like Query except it panics on error.
func (sq *Sqinn) MustQuery(sql string, values []interface{}, colTypes []ValueType) []Row {
	rows, err := sq.Query(sql, values, colTypes)
	if err != nil {
		panic(err)
	}
	return rows
}

func (sq *Sqinn) writeAndRead(req []byte) ([]byte, error) {
	// write req
	sz := len(req)
	buf := make([]byte, 0, 4+len(req))
	buf = append(buf, encodeInt32(sz)...)
	buf = append(buf, req...)
	if sq.logBinary {
		log.Printf("write 4 bytes req size: %v", buf[0:4])
		log.Printf("write %d bytes req payload: %v", len(req), req)
	}
	_, err := sq.sin.Write(buf)
	if err != nil {
		return nil, err
	}
	// read resp
	if sq.logBinary {
		log.Printf("waiting for 4 bytes resp size")
	}
	buf = make([]byte, 4)
	_, err = io.ReadFull(sq.sout, buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read resp size: %w", err)
	}
	if sq.logBinary {
		log.Printf("received %d bytes resp size: %v", len(buf), buf)
	}
	sz, _, err = decodeInt32(buf)
	if err != nil {
		return nil, err
	}
	if sz <= 0 {
		return nil, fmt.Errorf("invalid resp size %d", sz)
	}
	buf = make([]byte, sz)
	if sq.logBinary {
		log.Printf("waiting for %d resp payload", sz)
	}
	_, err = io.ReadFull(sq.sout, buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read resp payload: %w", err)
	}
	if sq.logBinary {
		log.Printf("received %d bytes resp payload: %v", len(buf), buf)
	}
	var success bool
	success, buf, err = decodeBool(buf)
	if err != nil {
		return nil, err
	}
	if !success {
		msg, _, err := decodeString(buf)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("sqinn: %s", msg)
	}
	return buf, nil
}

// Terminate terminates a running Sqinn instance.
// Each launched Sqinn instance should be terminated
// with Terminate. After Terminate has been called, this Sqinn
// instance must not be used interface{} more.
func (sq *Sqinn) Terminate() error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// a request of length zero makes sqinn terminate
	_, err := sq.sin.Write(encodeInt32(0))
	if err != nil {
		return err
	}
	err = sq.cmd.Wait()
	if err != nil {
		return err
	}
	sq.serr.Close()
	sq.sout.Close()
	sq.sin.Close()
	return nil
}
