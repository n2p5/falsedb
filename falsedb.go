package falsedb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
)

// DriverName is the name of the driver name used by sql
const DriverName = "falsedb"

// Driver conforms to the driver.Driver Interface
type Driver struct{}

// conn conforms to the driver.Conn interface
type conn struct{}

// stmt conforms to the driver.Stmt interface
type stmt struct{}

// result conforms to the driver.Result interface
type result struct{}

// value conforms to the driver.Value interface
type value struct{}

// rows conforms to the driver.Rows interface
type rows struct{}

// tx conforms to the driver.Tx interface
type tx struct{}

func init() {
	sql.Register("falsedb", &Driver{})
}

// Rollback conforms to the driver.Tx interface. Always returns nil.
func (t *tx) Commit() error { return nil }

// Rollback conforms to the driver.Tx interface. Always returns nil.
func (t *tx) Rollback() error { return nil }

// Columns conforms to the driver.Rows interface. Always returns an empty string slice.
func (r *rows) Columns() []string { return []string{} }

// Close conforms to the driver.Rows interface. Always returns nil.
func (r *rows) Close() error { return nil }

// Next conforms to the driver.Rows interface. Always returns nil.
func (r *rows) Next(dest []driver.Value) error { return io.EOF }

// LastInsertId conforms to the driver.Result interface
func (r *result) LastInsertId() (int64, error) { return 0, nil }

// RowsAffected conforms to the driver.Result interface
func (r *result) RowsAffected() (int64, error) { return 0, nil }

// Close conforms to the driver.Stmt interface. Always returns nil.
func (s *stmt) Close() error { return nil }

//  NumInput conforms to the driver.Stmt interface. Always returns 0.
func (s *stmt) NumInput() int { return 0 }

// Exec conforms to the driver.Stmt interface
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) { return &result{}, nil }

// Query conforms to the driver.Stmt interface
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) { return &rows{}, nil }

// Exec conforms to the driver.Conn interface.
func (c *conn) Prepare(query string) (driver.Stmt, error) { return &stmt{}, nil }

// Close conforms to the driver.Conn interface. Always returns nil.
func (c *conn) Close() error { return nil }

// Begin conforms to the driver.Conn interface.
func (c *conn) Begin() (driver.Tx, error) { return &tx{}, nil }

// Query conforms to the driver.Queryer interface
func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) { return &rows{}, nil }

// Exec conforms to the driver.Execer interface
func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) { return &result{}, nil }

// QueryContext conforms to the driver
func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return &rows{}, nil
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return &result{}, nil
}

// Open conforms to the Driver Interface, any string is valid
func (d *Driver) Open(connStr string) (driver.Conn, error) { return &conn{}, nil }

// Open is an easy way to Open a DB
func Open() (*sql.DB, error) { return sql.Open(DriverName, "") }

// NewDriver returns a pointer to a Driver
func NewDriver() *Driver { return &Driver{} }
