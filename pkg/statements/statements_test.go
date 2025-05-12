package statements

import (
	"errors"
	"io"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func MustClose(c io.Closer) {
	_ = c.Close()
}

func TestPrepare(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db)

	mock.ExpectPrepare("one")
	mock.ExpectPrepare("two")
	mock.ExpectPrepare("three")

	_, err = Prepare(db, "one", "two", "three")

	wantErr := false
	if (err != nil) != wantErr {
		t.Errorf("Prepare() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestPrepare_WithError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db)

	mock.ExpectPrepare("one").WillReturnError(errors.New("error"))

	_, err = Prepare(db, "one", "two", "three")

	wantErr := true
	if (err != nil) != wantErr {
		t.Errorf("Prepare() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestStatements_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db)

	mock.ExpectPrepare("one")
	mock.ExpectPrepare("two")
	mock.ExpectPrepare("three")

	stmts, err := Prepare(db, "one", "two", "three")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}

	if s := stmts.Get("one"); s == nil {
		t.Errorf("statements.Get(\"one\") should return a non-nil value")
	}
	if s := stmts.Get("two"); s == nil {
		t.Errorf("statements.Get(\"two\") should return a non-nil value")
	}
	if s := stmts.Get("three"); s == nil {
		t.Errorf("statements.Get(\"three\") should return a non-nil value")
	}
}
