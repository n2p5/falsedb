package falsedb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"
	"time"
)

func TestDriverRegistration(t *testing.T) {
	db, err := sql.Open(DriverName, "")
	if err != nil {
		t.Errorf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestOpenConnector(t *testing.T) {
	d := &Driver{}
	connector, err := d.OpenConnector("")
	if err != nil {
		t.Errorf("OpenConnector failed: %v", err)
	}

	conn, err := connector.Connect(context.Background())
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}
	defer conn.Close()

	if driver := connector.Driver(); driver == nil {
		t.Error("Driver() returned nil")
	}
}

func TestDirectDriver(t *testing.T) {
	// Test direct driver operations without sql.DB wrapper
	d := &Driver{}
	conn, err := d.Open("")
	if err != nil {
		t.Fatalf("Driver.Open failed: %v", err)
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// These should properly handle cancelled context
	if _, err := conn.(driver.ConnBeginTx).BeginTx(ctx, driver.TxOptions{}); err == nil {
		t.Error("Expected error for BeginTx with cancelled context")
	}

	if _, err := conn.(driver.ConnPrepareContext).PrepareContext(ctx, "SELECT 1"); err == nil {
		t.Error("Expected error for PrepareContext with cancelled context")
	}

	if _, err := conn.(driver.QueryerContext).QueryContext(ctx, "SELECT 1", nil); err == nil {
		t.Error("Expected error for QueryContext with cancelled context")
	}

	if _, err := conn.(driver.ExecerContext).ExecContext(ctx, "INSERT 1", nil); err == nil {
		t.Error("Expected error for ExecContext with cancelled context")
	}
}

func TestConnection(t *testing.T) {
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Test basic operations with live context
	ctx := context.Background()

	// Test query
	rows, err := db.QueryContext(ctx, "SELECT * FROM anything")
	if err != nil {
		t.Errorf("QueryContext failed: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		t.Error("Expected no rows, but got one")
	}

	// Test exec
	result, err := db.ExecContext(ctx, "INSERT INTO void VALUES (?)", 42)
	if err != nil {
		t.Errorf("ExecContext failed: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil || id != 0 {
		t.Errorf("Expected LastInsertId=0, got %d with error: %v", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		t.Errorf("Expected RowsAffected=0, got %d with error: %v", affected, err)
	}
}

func TestTransactions(t *testing.T) {
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test successful transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("BeginTx failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("Commit failed: %v", err)
	}

	// Test rollback
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("BeginTx failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Errorf("Rollback failed: %v", err)
	}

	// Since sql.DB enforces context cancellation, we test the driver directly
	d := &Driver{}
	conn, _ := d.Open("")
	connTx, _ := conn.(driver.ConnBeginTx)

	// This should return error due to cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := connTx.BeginTx(ctx, driver.TxOptions{}); err == nil {
		t.Error("Expected error for BeginTx with cancelled context at driver level")
	}
}

func TestPreparedStatements(t *testing.T) {
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test prepare with live context
	stmt, err := db.PrepareContext(ctx, "SELECT * FROM void WHERE id = ?")
	if err != nil {
		t.Errorf("PrepareContext failed: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, 42)
	if err != nil {
		t.Errorf("QueryContext on prepared statement failed: %v", err)
	}
	defer rows.Close()

	// Test with driver directly for context cancellation
	d := &Driver{}
	conn, _ := d.Open("")
	connPrep, _ := conn.(driver.ConnPrepareContext)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := connPrep.PrepareContext(ctx, "SELECT 1"); err == nil {
		t.Error("Expected error for PrepareContext with cancelled context at driver level")
	}
}

func TestRows(t *testing.T) {
	r := &rows{}

	if cols := r.Columns(); len(cols) != 0 {
		t.Errorf("Expected empty columns, got %v", cols)
	}

	if err := r.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if err := r.Next(nil); err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

func TestRowsColumnTypes(t *testing.T) {
	r := &rows{}

	if typ := r.ColumnTypeScanType(0); typ != nil {
		t.Errorf("Expected nil scan type, got %v", typ)
	}

	if name := r.ColumnTypeDatabaseTypeName(0); name != "NULL" {
		t.Errorf("Expected NULL type name, got %s", name)
	}

	if nullable, ok := r.ColumnTypeNullable(0); !nullable || !ok {
		t.Errorf("Expected nullable=true, ok=true, got nullable=%v, ok=%v", nullable, ok)
	}

	if precision, scale, ok := r.ColumnTypePrecisionScale(0); precision != 0 || scale != 0 || !ok {
		t.Errorf("Expected precision=0, scale=0, ok=true, got precision=%d, scale=%d, ok=%v",
			precision, scale, ok)
	}

	if length, ok := r.ColumnTypeLength(0); length != 0 || !ok {
		t.Errorf("Expected length=0, ok=true, got length=%d, ok=%v", length, ok)
	}
}

func TestConfig(t *testing.T) {
	maxOpen := 10
	maxIdle := 5
	lifetime := time.Hour
	idleTime := time.Minute

	db := OpenDB(
		WithMaxOpenConns(maxOpen),
		WithMaxIdleConns(maxIdle),
		WithConnMaxLifetime(lifetime),
		WithConnMaxIdleTime(idleTime),
	)
	defer db.Close()

	if got := db.Stats().MaxOpenConnections; got != maxOpen {
		t.Errorf("Expected MaxOpenConnections=%d, got %d", maxOpen, got)
	}
}

func TestQueryContext(t *testing.T) {
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test with live context
	rows, err := db.QueryContext(ctx, "SELECT * FROM void WHERE id = ?", 42)
	if err != nil {
		t.Errorf("QueryContext failed: %v", err)
	}
	defer rows.Close()

	// Test directly with driver for cancelled context
	d := &Driver{}
	conn, _ := d.Open("")
	connQuery, _ := conn.(driver.QueryerContext)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := connQuery.QueryContext(ctx, "SELECT 1", nil); err == nil {
		t.Error("Expected error for QueryContext with cancelled context at driver level")
	}
}

func TestExecContext(t *testing.T) {
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test with live context
	result, err := db.ExecContext(ctx, "INSERT INTO void VALUES (?)", 42)
	if err != nil {
		t.Errorf("ExecContext failed: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		t.Errorf("Expected RowsAffected=0, got %d with error: %v", affected, err)
	}

	// Test directly with driver for cancelled context
	d := &Driver{}
	conn, _ := d.Open("")
	connExec, _ := conn.(driver.ExecerContext)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := connExec.ExecContext(ctx, "INSERT 1", nil); err == nil {
		t.Error("Expected error for ExecContext with cancelled context at driver level")
	}
}

func TestNamedValue(t *testing.T) {
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test with named parameters
	result, err := db.ExecContext(ctx, "INSERT INTO void VALUES (:value)",
		sql.Named("value", 42))
	if err != nil {
		t.Errorf("ExecContext with named values failed: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		t.Errorf("Expected RowsAffected=0, got %d with error: %v", affected, err)
	}

	// Test query with named parameters
	rows, err := db.QueryContext(ctx, "SELECT * FROM void WHERE id = :id",
		sql.Named("id", 42))
	if err != nil {
		t.Errorf("QueryContext with named values failed: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		t.Error("Expected no rows from named parameter query")
	}
}

func TestConnector(t *testing.T) {
	db := OpenDB()
	defer db.Close()

	connector, err := (&Driver{}).OpenConnector("")
	if err != nil {
		t.Fatalf("OpenConnector failed: %v", err)
	}

	conn, err := connector.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer conn.Close()

	// Test statement on connection
	stmt, err := conn.(driver.Conn).Prepare("SELECT 1")
	if err != nil {
		t.Errorf("Prepare on connector's connection failed: %v", err)
	}
	defer stmt.Close()

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	conn2, err := connector.Connect(ctx)
	if err != nil {
		t.Errorf("Connect with cancelled context failed: %v", err)
	}
	defer conn2.Close()
}

func TestStmtContext(t *testing.T) {
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Prepare statement
	stmt, err := db.PrepareContext(ctx, "SELECT * FROM void WHERE id = ?")
	if err != nil {
		t.Fatalf("PrepareContext failed: %v", err)
	}
	defer stmt.Close()

	// Test ExecContext on statement
	result, err := stmt.ExecContext(ctx, 42)
	if err != nil {
		t.Errorf("Statement ExecContext failed: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected != 0 {
		t.Errorf("Expected RowsAffected=0, got %d with error: %v", affected, err)
	}

	// Test QueryContext on statement
	rows, err := stmt.QueryContext(ctx, 42)
	if err != nil {
		t.Errorf("Statement QueryContext failed: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		t.Error("Expected no rows from statement query")
	}

	// Test directly with driver for cancelled context
	d := &Driver{}
	conn, _ := d.Open("")
	driverStmt, _ := conn.(driver.Conn).Prepare("SELECT 1")
	stmtCtx, _ := driverStmt.(driver.StmtExecContext)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := stmtCtx.ExecContext(ctx, nil); err == nil {
		t.Error("Expected error for Statement ExecContext with cancelled context at driver level")
	}
}
