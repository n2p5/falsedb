package falsedb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"time"
)

// DriverName is the name of the driver used by sql
const DriverName = "falsedb"

func init() {
	sql.Register(DriverName, &Driver{})
}

// Driver implements database/sql/driver.Driver and driver.DriverContext
type Driver struct{}

func (d *Driver) Open(name string) (driver.Conn, error) {
	return &conn{}, nil
}

func (d *Driver) OpenConnector(name string) (driver.Connector, error) {
	return &connector{driver: d}, nil
}

type connector struct {
	driver *Driver
}

func (c *connector) Connect(context.Context) (driver.Conn, error) {
	return c.driver.Open("")
}

func (c *connector) Driver() driver.Driver {
	return c.driver
}

// conn implements database/sql/driver.Conn, driver.ConnBeginTx, driver.QueryerContext,
// driver.ExecerContext, and driver.ConnPrepareContext
type conn struct{}

func (c *conn) Begin() (driver.Tx, error) {
	return &tx{}, nil
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{numInput: -1}, nil
}

// Direct driver methods ignore context
func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &tx{}, nil
	}
}

func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &stmt{numInput: -1}, nil
	}
}

func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &rows{}, nil
	}
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &result{}, nil
	}
}

// Legacy methods for driver.Queryer and driver.Execer
func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return &rows{}, nil
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return &result{}, nil
}

type stmt struct {
	numInput int
}

func (s *stmt) Close() error {
	return nil
}

func (s *stmt) NumInput() int {
	return s.numInput
}

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return &result{}, nil
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return &rows{}, nil
}

// StmtExecContext and StmtQueryContext implementations
func (s *stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &result{}, nil
	}
}

func (s *stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &rows{}, nil
	}
}

type result struct{}

func (r *result) LastInsertId() (int64, error) {
	return 0, nil
}

func (r *result) RowsAffected() (int64, error) {
	return 0, nil
}

type rows struct {
	closed bool
}

func (r *rows) Columns() []string {
	return []string{}
}

func (r *rows) Close() error {
	r.closed = true
	return nil
}

func (r *rows) Next(dest []driver.Value) error {
	return io.EOF
}

// Optional column type interfaces
func (r *rows) ColumnTypeScanType(index int) any {
	return nil
}

func (r *rows) ColumnTypeDatabaseTypeName(index int) string {
	return "NULL"
}

func (r *rows) ColumnTypeNullable(index int) (nullable, ok bool) {
	return true, true
}

func (r *rows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return 0, 0, true
}

func (r *rows) ColumnTypeLength(index int) (length int64, ok bool) {
	return 0, true
}

type tx struct{}

func (t *tx) Commit() error {
	return nil
}

func (t *tx) Rollback() error {
	return nil
}

// Helper functions for opening connections
func Open() (*sql.DB, error) {
	return sql.Open(DriverName, "")
}

func OpenDB(opts ...Option) *sql.DB {
	cfg := &config{
		maxOpenConns:    0,
		maxIdleConns:    0,
		connMaxLifetime: 0,
		connMaxIdleTime: 0,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	db, _ := Open()
	db.SetMaxOpenConns(cfg.maxOpenConns)
	db.SetMaxIdleConns(cfg.maxIdleConns)
	db.SetConnMaxLifetime(cfg.connMaxLifetime)
	db.SetConnMaxIdleTime(cfg.connMaxIdleTime)
	return db
}

// Configuration options
type Option func(*config)

type config struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
}

func WithMaxOpenConns(n int) Option {
	return func(c *config) {
		c.maxOpenConns = n
	}
}

func WithMaxIdleConns(n int) Option {
	return func(c *config) {
		c.maxIdleConns = n
	}
}

func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *config) {
		c.connMaxLifetime = d
	}
}

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *config) {
		c.connMaxIdleTime = d
	}
}
