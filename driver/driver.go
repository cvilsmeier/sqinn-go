package driver

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"net/url"
	"os"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

var driverName = "sqlite3"

type sqlite int8

func init() {
	if driverName != "" {
		sql.Register(driverName, sqlite(0))
	}
}

func (d sqlite) Open(name string) (driver.Conn, error) {
	tokens, err := url.Parse(name)
	if err != nil {
		return nil, err
	}
	sqinn, err := sqinn.Launch(sqinn.Options{
		SqinnPath: tokens.Query().Get("sqinnpath"),
	})
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(tokens.Path); os.IsNotExist(err) {
		return nil, err
	}

	sqinn.Open(tokens.Path)
	return &connection{sqinn: sqinn}, nil
}

type connection struct {
	sqinn *sqinn.Sqinn
}

func (c *connection) Prepare(query string) (driver.Stmt, error) {
	err := c.sqinn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &statement{sqinn: c.sqinn}, nil
}

func (c *connection) Close() error {
	if err := c.sqinn.Close(); err != nil {
		return err
	}
	return c.sqinn.Terminate()
}

func (c *connection) Begin() (driver.Tx, error) {
	return nil, nil
}

type result struct {
	lastInsertId int64
	rowsAffected int64
}

func (r *result) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type statement struct {
	sqinn *sqinn.Sqinn
}

func (s *statement) Close() error {
	return s.sqinn.Finalize()
}

func (s *statement) NumInput() int {
	return -1
}

func (s *statement) Exec(args []driver.Value) (driver.Result, error) {
	for i, v := range args {
		if err := s.sqinn.Bind(i, v); err != nil {
			return nil, err
		}
	}
	if _, err := s.sqinn.Step(); err != nil {
		return nil, err
	}

	res := result{
		lastInsertId: -1,
	}

	changes, err := s.sqinn.Changes()
	if err != nil {
		changes = -1
	}
	res.rowsAffected = int64(changes)

	// TODO: implement last_insert_rowid

	return &res, nil
}

func (s *statement) Query(args []driver.Value) (driver.Rows, error) {
	res := rows{
		sqinn: s.sqinn,
	}

	for i, v := range args {
		if err := s.sqinn.Bind(i, v); err != nil {
			return nil, err
		}
	}

	more, err := s.sqinn.Step()
	if err != nil {
		return nil, err
	}
	res.lastRowReached = !more

	count, err := s.sqinn.ColumnCount()
	if err != nil {
		return nil, err
	}
	for i := 0; i < count; i++ {
		colType, err := s.sqinn.ColumnType(i)
		if err != nil {
			return nil, err
		}
		colName, err := s.sqinn.ColumnName(i)
		if err != nil {
			return nil, err
		}

		res.columns = append(res.columns, columnInfo{
			kind: colType,
			name: colName,
		})
	}

	return &res, err
}

type columnInfo struct {
	kind sqinn.ValueType
	name string
}

type rows struct {
	sqinn          *sqinn.Sqinn
	columns        []columnInfo
	lastRowReached bool
}

func (r *rows) Columns() []string {
	result := make([]string, len(r.columns))
	for i, v := range r.columns {
		result[i] = v.name
	}
	return result
}

func (r *rows) Close() error {
	return r.sqinn.Finalize()
}

func (r *rows) Next(dest []driver.Value) error {
	for i, info := range r.columns {
		val, err := r.sqinn.Column(i, info.kind)
		if err != nil {
			return err
		}
		dest[i] = val.AsValue(info.kind)
	}

	if r.lastRowReached {
		return io.EOF
	}

	more, err := r.sqinn.Step()
	if err != nil {
		return err
	}
	r.lastRowReached = !more

	return nil
}
