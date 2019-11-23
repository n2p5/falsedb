package falsedb

import (
	"testing"
)

func TestOpen(t *testing.T) {
	t.Parallel()
	db, err := Open()
	if err != nil {
		t.Error(err)
	}
	if err := db.Close(); err != nil {
		t.Error(err)
	}
}

func TestDriverOpen(t *testing.T) {
	t.Parallel()
	d := NewDriver()
	conn, err := d.Open(DriverName)
	if err != nil {
		t.Error(err)
	}
	if err := conn.Close(); err != nil {
		t.Error(err)
	}
}

func TestGetRows(t *testing.T) {
	// t.Parallel()
	db, err := Open()
	if err != nil {
		t.Error(err)
	}
	rows, err := db.Query("SELECT * from lulz")
	if err != nil {
		t.Error(err)
	}
	for rows.Next() {
		t.Error("FalseDB is always empty. this should never happen")
	}
}
