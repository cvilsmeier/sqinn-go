package sqinn

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"sync"
)

// Logger logs sqinn error and debug messages
type Logger interface {
	Log(s string)
}

// StdLogger logs to a stdlib log.Logger or to log.DefaultLogger.
type StdLogger struct {
	Logger *log.Logger
}

func (l StdLogger) Log(s string) {
	if l.Logger != nil {
		l.Logger.Println(s)
	} else {
		log.Println(s)
	}
}

// NoLogger does not log anything.
type NoLogger struct{}

func (l NoLogger) Log(s string) {}

// function codes, see sqinn/src/handler.h

const (
	FC_SQINN_VERSION  byte = 1
	FC_IO_VERSION     byte = 2
	FC_SQLITE_VERSION byte = 3
	FC_OPEN           byte = 10
	FC_PREPARE        byte = 11
	FC_BIND           byte = 12
	FC_STEP           byte = 13
	FC_RESET          byte = 14
	FC_CHANGES        byte = 15
	FC_COLUMN         byte = 16
	FC_FINALIZE       byte = 17
	FC_CLOSE          byte = 18
	FC_EXEC           byte = 51
	FC_QUERY          byte = 52
)

// value types, see sqinn/src/handler.h

const (
	VAL_NULL   byte = 0
	VAL_INT    byte = 1
	VAL_INT64  byte = 2
	VAL_DOUBLE byte = 3
	VAL_TEXT   byte = 4
	VAL_BLOB   byte = 5
)

// SQL values can be null and therefore must be wrapped

type IntValue struct {
	Set   bool
	Value int
}

type Int64Value struct {
	Set   bool
	Value int64
}

type DoubleValue struct {
	Set   bool
	Value float64
}

type StringValue struct {
	Set   bool
	Value string
}

type BlobValue struct {
	Set   bool
	Value []byte
}

type AnyValue struct {
	Int    IntValue
	Int64  Int64Value
	Double DoubleValue
	String StringValue
	Blob   BlobValue
}

func (a AnyValue) AsInt() int {
	return a.Int.Value
}

func (a AnyValue) AsInt64() int64 {
	return a.Int64.Value
}

func (a AnyValue) AsDouble() float64 {
	return a.Double.Value
}

func (a AnyValue) AsString() string {
	return a.String.Value
}

func (a AnyValue) AsBlob() []byte {
	return a.Blob.Value
}

type Row struct {
	Values []AnyValue
}

// marshalling

func encodeInt32(v int) []byte {
	return []byte{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v >> 0),
	}
}

func encodeString(v string) []byte {
	data := []byte(v)
	sz := len(data) + 1
	buf := make([]byte, 0, 4+sz)
	buf = append(buf, encodeInt32(sz)...)
	buf = append(buf, data...)
	buf = append(buf, 0)
	return buf
}

func encodeBlob(v []byte) []byte {
	sz := len(v)
	buf := make([]byte, 0, 4+sz)
	buf = append(buf, encodeInt32(sz)...)
	buf = append(buf, v...)
	return buf
}

func encodeDouble(v float64) []byte {
	s := strconv.FormatFloat(v, 'g', -1, 64)
	return encodeString(s)
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

func decodeString(buf []byte) (string, []byte) {
	sz, buf := decodeInt32(buf)
	if len(buf) < sz {
		panic(fmt.Errorf("cannot decodeString length %d from a %d byte buffer", sz, len(buf)))
	}
	v := string(buf[:sz-1])
	buf = buf[sz:]
	return v, buf
}

func decodeBlob(buf []byte) ([]byte, []byte) {
	sz, buf := decodeInt32(buf)
	if len(buf) < sz {
		panic(fmt.Errorf("cannot decodeBlob length %d from a %d byte buffer", sz, len(buf)))
	}
	v := buf[:sz]
	buf = buf[sz:]
	return v, buf
}

func decodeDouble(buf []byte) (float64, []byte) {
	s, buf := decodeString(buf)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(fmt.Errorf("cannot decodeDouble: %s", err))
	}
	return v, buf
}

func decodeBool(buf []byte) (bool, []byte) {
	if len(buf) < 1 {
		panic(fmt.Errorf("cannot decodeBool from a %d byte buffer", len(buf)))
	}
	v := buf[0] != 0
	buf = buf[1:]
	return v, buf
}

type Options struct {
	SqinnPath string
	Logger    Logger
}

type Sqinn struct {
	mx   sync.Mutex
	cmd  *exec.Cmd
	sin  io.WriteCloser
	sout io.ReadCloser
	serr io.ReadCloser
}

func New(options Options) (*Sqinn, error) {
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
	sq := &Sqinn{sync.Mutex{}, cmd, sin, sout, serr}
	logger := options.Logger
	if logger == nil {
		logger = NoLogger{}
	}
	go sq.run(logger)
	return sq, nil
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

func (sq *Sqinn) SqinnVersion(filename string) (string, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{FC_SQINN_VERSION}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return "", err
	}
	var version string
	version, resp = decodeString(resp)
	return version, nil
}

func (sq *Sqinn) IoVersion(filename string) (int, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{FC_IO_VERSION}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return 0, err
	}
	var version int
	version, resp = decodeInt32(resp)
	return version, nil
}

func (sq *Sqinn) SqliteVersion(filename string) (string, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{FC_SQLITE_VERSION}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return "", err
	}
	var version string
	version, resp = decodeString(resp)
	return version, nil
}

func (sq *Sqinn) Open(filename string) error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 10+len(filename))
	req = append(req, FC_OPEN)
	req = append(req, encodeString(filename)...)
	// resp
	_, err := sq.writeAndRead(req)
	if err != nil {
		return err
	}
	return nil
}

func (sq *Sqinn) Prepare(sql string) error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 10+len(sql))
	req = append(req, FC_PREPARE)
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
		req = append(req, VAL_NULL)
	case int:
		req = append(req, VAL_INT)
		req = append(req, encodeInt32(v)...)
	case float64:
		req = append(req, VAL_DOUBLE)
		req = append(req, encodeDouble(float64(v))...)
	case string:
		req = append(req, VAL_TEXT)
		req = append(req, encodeString(v)...)
	case []byte:
		req = append(req, VAL_BLOB)
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

func (sq *Sqinn) Bind(iparam int, value interface{}) error {
	if iparam < 1 {
		return fmt.Errorf("Bind: iparam must be >= 1 but was %d", iparam)
	}
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 6)
	req = append(req, FC_BIND)
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

func (sq *Sqinn) Step() (bool, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{FC_STEP}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return false, err
	}
	more, _ := decodeBool(resp)
	return more, nil
}

func (sq *Sqinn) Reset() error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{FC_RESET}
	// resp
	_, err := sq.writeAndRead(req)
	if err != nil {
		return err
	}
	return nil
}

func (sq *Sqinn) Changes() (int, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{FC_CHANGES}
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return 0, err
	}
	var changes int
	changes, resp = decodeInt32(resp)
	return changes, nil
}

func (sq *Sqinn) ColumnInt(icol int) (IntValue, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 6)
	req = append(req, FC_COLUMN)
	req = append(req, encodeInt32(icol)...)
	req = append(req, VAL_INT)
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return IntValue{}, err
	}
	set, resp := decodeBool(resp)
	if !set {
		return IntValue{}, nil
	}
	v, _ := decodeInt32(resp)
	return IntValue{Set: true, Value: v}, nil
}

func (sq *Sqinn) ColumnDouble(icol int) (DoubleValue, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 6)
	req = append(req, FC_COLUMN)
	req = append(req, encodeInt32(icol)...)
	req = append(req, VAL_DOUBLE)
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return DoubleValue{}, err
	}
	set, resp := decodeBool(resp)
	if !set {
		return DoubleValue{}, nil
	}
	str, _ := decodeString(resp)
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return DoubleValue{}, err
	}
	return DoubleValue{Set: true, Value: v}, nil
}

func (sq *Sqinn) ColumnText(icol int) (StringValue, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := make([]byte, 0, 6)
	req = append(req, FC_COLUMN)
	req = append(req, encodeInt32(icol)...)
	req = append(req, VAL_TEXT)
	// resp
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return StringValue{}, err
	}
	set, resp := decodeBool(resp)
	if !set {
		return StringValue{}, nil
	}
	v, _ := decodeString(resp)
	return StringValue{Set: true, Value: v}, nil
}

func (sq *Sqinn) Finalize() error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{FC_FINALIZE}
	// resp
	_, err := sq.writeAndRead(req)
	return err
}

func (sq *Sqinn) Close() error {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	// req
	req := []byte{FC_CLOSE}
	// resp
	_, err := sq.writeAndRead(req)
	return err
}

func (sq *Sqinn) ExecOne(sql string) (int, error) {
	changes, err := sq.Exec(sql, 1, 0, nil)
	if err != nil {
		return 0, err
	}
	return changes[0], nil
}

func (sq *Sqinn) Exec(sql string, niterations, nparams int, values []interface{}) ([]int, error) {
	if niterations < 0 {
		return nil, fmt.Errorf("Exec '%s' niterations must be >= 0 but was %d", sql, niterations)
	}
	if len(values) != niterations*nparams {
		return nil, fmt.Errorf("Exec '%s' expected %d values but have %d", sql, niterations*nparams, len(values))
	}
	sq.mx.Lock()
	defer sq.mx.Unlock()
	req := make([]byte, 0, 10+len(sql))
	req = append(req, FC_EXEC)
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
		changes[i], resp = decodeInt32(resp)
	}
	return changes, nil
}

func (sq *Sqinn) Query(sql string, values []interface{}, colTypes []byte) ([]Row, error) {
	sq.mx.Lock()
	defer sq.mx.Unlock()
	req := make([]byte, 0, 10+len(sql))
	req = append(req, FC_QUERY)
	req = append(req, encodeString(sql)...)
	nparams := len(values)
	req = append(req, encodeInt32(nparams)...)
	var err error
	req, err = sq.bindValues(req, values)
	ncols := len(colTypes)
	req = append(req, encodeInt32(ncols)...)
	req = append(req, colTypes...)
	resp, err := sq.writeAndRead(req)
	if err != nil {
		return nil, err
	}
	var nrows int
	nrows, resp = decodeInt32(resp)
	rows := make([]Row, 0, nrows)
	for i := 0; i < nrows; i++ {
		var row Row
		row.Values = make([]AnyValue, 0, ncols)
		for icol := 0; icol < ncols; icol++ {
			var any AnyValue
			var set bool
			set, resp = decodeBool(resp)
			if set {
				switch colTypes[icol] {
				case VAL_INT:
					any.Int.Set = true
					any.Int.Value, resp = decodeInt32(resp)
				case VAL_DOUBLE:
					any.Double.Set = true
					any.Double.Value, resp = decodeDouble(resp)
				case VAL_TEXT:
					any.String.Set = true
					any.String.Value, resp = decodeString(resp)
				case VAL_BLOB:
					any.Blob.Set = true
					any.Blob.Value, resp = decodeBlob(resp)
				default:
					return nil, fmt.Errorf("invalid col type %d", colTypes[icol])
				}
			}
			row.Values = append(row.Values, any)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (sq *Sqinn) writeAndRead(req []byte) ([]byte, error) {
	traceReq := false
	traceResp := false
	// write req
	sz := len(req)
	buf := make([]byte, 0, 4+len(req))
	buf = append(buf, encodeInt32(sz)...)
	buf = append(buf, req...)
	if traceReq {
		log.Printf("write %d bytes sz+req: %v", len(buf), buf)
	}
	_, err := sq.sin.Write(buf)
	if err != nil {
		return nil, err
	}
	// read resp
	if traceResp {
		// time.Sleep(100 * time.Millisecond)
		log.Printf("waiting for 4 bytes resp sz")
	}
	buf = make([]byte, 4)
	_, err = io.ReadFull(sq.sout, buf)
	if err != nil {
		return nil, fmt.Errorf("while reading from sqinn: %w", err)
	}
	if traceResp {
		log.Printf("received %d bytes resp length: %v", len(buf), buf)
	}
	sz, _ = decodeInt32(buf)
	if traceResp {
		log.Printf("resp length will be %d bytes", sz)
	}
	if sz <= 0 {
		return nil, fmt.Errorf("invalid response size %d", sz)
	}
	buf = make([]byte, sz)
	if traceResp {
		log.Printf("waiting for %d resp data", sz)
	}
	_, err = io.ReadFull(sq.sout, buf)
	if err != nil {
		return nil, fmt.Errorf("while reading from sqinn: %w", err)
	}
	if traceResp {
		log.Printf("received %d bytes resp data: %v", len(buf), buf)
		// time.Sleep(100 * time.Millisecond)
	}
	var ok bool
	ok, buf = decodeBool(buf)
	if !ok {
		msg, _ := decodeString(buf)
		return nil, fmt.Errorf("sqinn: %s", msg)
	}
	return buf, nil
}

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
