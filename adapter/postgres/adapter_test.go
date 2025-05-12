package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/mertenvg/migrate/pkg/reader"
)

func TestAdapter_Begin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectBegin()

	a := NewAdapter(db)

	wantErr := false
	if err := a.Begin(context.Background()); (err != nil) != wantErr {
		t.Errorf("Begin() error = %v, wantErr %v", err, wantErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Begin_WithNilContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectBegin()

	a := NewAdapter(db)

	wantErr := false
	if err := a.Begin(nil); (err != nil) != wantErr {
		t.Errorf("Begin() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Begin_WithError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	a := NewAdapter(db)

	wantErr := true
	if err := a.Begin(context.Background()); (err != nil) != wantErr {
		t.Errorf("Begin() error = %v, wantErr %v", err, wantErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectQuery(makeMockFriendly(migrations)).WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("aaa").AddRow("bbb"),
	)

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false
	want := []string{"aaa", "bbb"}

	got, err := a.List()
	if (err != nil) != wantErr {
		t.Errorf("List() error = %v, wantErr %v", err, wantErr)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_List_WithQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectQuery(makeMockFriendly(migrations)).WillReturnError(errors.New("query error"))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true
	var want []string

	got, err := a.List()
	if (err != nil) != wantErr {
		t.Errorf("List() error = %v, wantErr %v", err, wantErr)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_List_WithScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectQuery(makeMockFriendly(migrations)).WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow(nil),
	)

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true
	var want []string

	got, err := a.List()
	if (err != nil) != wantErr {
		t.Errorf("List() error = %v, wantErr %v", err, wantErr)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Down(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectQuery(makeMockFriendly(rollbackWithName)).WithArgs("aaa").WillReturnRows(
		sqlmock.NewRows([]string{"rollback"}).AddRow("rollback aaa"),
	)
	mock.ExpectExec(makeMockFriendly("rollback aaa")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(makeMockFriendly(removeWithName)).WithArgs("aaa").WillReturnResult(sqlmock.NewResult(0, 1))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false

	err = a.Down("aaa")
	if (err != nil) != wantErr {
		t.Errorf("Down() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Down_WithQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectQuery(makeMockFriendly(rollbackWithName)).WithArgs("aaa").WillReturnError(errors.New("query error"))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Down("aaa")
	if (err != nil) != wantErr {
		t.Errorf("Down() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Down_WithEmptyRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectQuery(makeMockFriendly(rollbackWithName)).WithArgs("aaa").WillReturnRows(
		sqlmock.NewRows([]string{"rollback"}).AddRow(""),
	)

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false

	err = a.Down("aaa")
	if (err != nil) != wantErr {
		t.Errorf("Down() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Down_WithSQLReaderFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectQuery(makeMockFriendly(rollbackWithName)).WithArgs("aaa").WillReturnRows(
		sqlmock.NewRows([]string{"rollback"}).AddRow("rollback aaa"),
	)
	// mock.ExpectExec(makeMockFriendly("rollback aaa")).WillReturnResult(sqlmock.NewResult(0, 0))
	// mock.ExpectExec(makeMockFriendly(removeWithName)).WithArgs("aaa").WillReturnResult(sqlmock.NewResult(0, 1))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	reader.FailNext = errors.New("fail next")

	err = a.Down("aaa")
	if (err != nil) != wantErr {
		t.Errorf("Down() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Down_WithBeginAndCommit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectQuery(makeMockFriendly(rollbackWithName)).WithArgs("aaa").WillReturnRows(
		sqlmock.NewRows([]string{"rollback"}).AddRow("begin; rollback aaa; commit;"),
	)
	mock.ExpectExec(makeMockFriendly("rollback aaa")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(makeMockFriendly(removeWithName)).WithArgs("aaa").WillReturnResult(sqlmock.NewResult(0, 1))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false

	err = a.Down("aaa")
	if (err != nil) != wantErr {
		t.Errorf("Down() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Down_WithQueryFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectQuery(makeMockFriendly(rollbackWithName)).WithArgs("aaa").WillReturnRows(
		sqlmock.NewRows([]string{"rollback"}).AddRow("begin; rollback aaa; commit;"),
	)
	mock.ExpectExec(makeMockFriendly("rollback aaa")).WillReturnError(errors.New("fail rollback aaa"))
	// mock.ExpectExec(makeMockFriendly(removeWithName)).WithArgs("aaa").WillReturnResult(sqlmock.NewResult(0, 1))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Down("aaa")
	if (err != nil) != wantErr {
		t.Errorf("Down() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Down_WithSaveStateFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectQuery(makeMockFriendly(rollbackWithName)).WithArgs("aaa").WillReturnRows(
		sqlmock.NewRows([]string{"rollback"}).AddRow("begin; rollback aaa; commit;"),
	)
	mock.ExpectExec(makeMockFriendly("rollback aaa")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(makeMockFriendly(removeWithName)).WithArgs("aaa").WillReturnError(errors.New("fail"))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Down("aaa")
	if (err != nil) != wantErr {
		t.Errorf("Down() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Up(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectExec(makeMockFriendly("apply aaa")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(makeMockFriendly(add)).WithArgs("aaa", "rollback aaa").WillReturnResult(sqlmock.NewResult(0, 1))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false

	err = a.Up("aaa", bytes.NewBufferString("apply aaa"), bytes.NewBufferString("rollback aaa"))
	if (err != nil) != wantErr {
		t.Errorf("Up() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Up_WithSQLReaderFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	reader.FailNext = errors.New("fail")

	err = a.Up("aaa", bytes.NewBufferString("apply aaa"), bytes.NewBufferString("rollback aaa"))
	if (err != nil) != wantErr {
		t.Errorf("Up() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Up_WithBeginAndCommit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectExec(makeMockFriendly("apply aaa")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(makeMockFriendly(add)).WithArgs("aaa", "rollback aaa").WillReturnResult(sqlmock.NewResult(0, 1))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false

	err = a.Up("aaa", bytes.NewBufferString("begin; apply aaa; commit;"), bytes.NewBufferString("rollback aaa"))
	if (err != nil) != wantErr {
		t.Errorf("Up() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Up_WithNilDownArg(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectExec(makeMockFriendly("apply aaa")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(makeMockFriendly(add)).WithArgs("aaa", "").WillReturnResult(sqlmock.NewResult(0, 1))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false

	err = a.Up("aaa", bytes.NewBufferString("begin; apply aaa; commit;"), nil)
	if (err != nil) != wantErr {
		t.Errorf("Up() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

type FailReader struct{}

func (r FailReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("fail")
}

func TestAdapter_Up_WithDownReadFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectExec(makeMockFriendly("apply aaa")).WillReturnResult(sqlmock.NewResult(0, 0))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Up("aaa", bytes.NewBufferString("begin; apply aaa; commit;"), &FailReader{})
	if (err != nil) != wantErr {
		t.Errorf("Up() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Up_WithSaveStateFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectExec(makeMockFriendly("apply aaa")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(makeMockFriendly(add)).WithArgs("aaa", "rollback aaa").WillReturnError(errors.New("fail"))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Up("aaa", bytes.NewBufferString("apply aaa"), bytes.NewBufferString("rollback aaa"))
	if (err != nil) != wantErr {
		t.Errorf("Up() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectCommit()

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false

	err = a.Commit()
	if (err != nil) != wantErr {
		t.Errorf("Commit() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Commit_WithoutTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Commit()
	if (err != nil) != wantErr {
		t.Errorf("Commit() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Commit_WithFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("fail"))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Commit()
	if (err != nil) != wantErr {
		t.Errorf("Commit() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectRollback()

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := false

	err = a.Rollback()
	if (err != nil) != wantErr {
		t.Errorf("Rollback() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Rollback_WithoutTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Rollback()
	if (err != nil) != wantErr {
		t.Errorf("Rollback() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Rollback_WithFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))
	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(errors.New("fail"))

	a := NewAdapter(db)
	err = a.Setup()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	err = a.Begin(nil)
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}

	wantErr := true

	err = a.Rollback()
	if (err != nil) != wantErr {
		t.Errorf("Rollback() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

var matchWhitespace = regexp.MustCompile("\\s+")

func makeMockFriendly(s string) string {
	return regexp.QuoteMeta(strings.TrimSpace(matchWhitespace.ReplaceAllString(s, " ")))
}

func TestAdapter_Setup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add))
	mock.ExpectPrepare(makeMockFriendly(migrations))
	mock.ExpectPrepare(makeMockFriendly(rollbackWithName))
	mock.ExpectPrepare(makeMockFriendly(removeWithName))

	a := NewAdapter(db)

	wantErr := false
	if err := a.Setup(); (err != nil) != wantErr {
		t.Errorf("Setup() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Setup_FailExec(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnError(errors.New("test error"))

	a := NewAdapter(db)

	wantErr := true
	if err := a.Setup(); (err != nil) != wantErr {
		t.Errorf("Setup() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAdapter_Setup_FailPrepare(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	mock.ExpectExec(makeMockFriendly(migrationStore)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(makeMockFriendly(add)).WillReturnError(errors.New("prepare error"))

	a := NewAdapter(db)

	wantErr := true
	if err := a.Setup(); (err != nil) != wantErr {
		t.Errorf("Setup() error = %v, wantErr %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

type MockCloser struct {
	closeCalled bool
}

func (c MockCloser) Close() error {
	c.closeCalled = true
	return fmt.Errorf("close error")
}

func TestMustClose(t *testing.T) {
	logFuncCalled := false
	logFunc := func(v ...any) { logFuncCalled = true }
	type args struct {
		c   MockCloser
		log LogFunc
	}
	tests := []struct {
		name    string
		args    args
		wantLog bool
	}{
		{
			name: "must close with log func",
			args: args{
				c:   MockCloser{},
				log: logFunc,
			},
			wantLog: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MustClose(tt.args.c, tt.args.log)
			if logFuncCalled != tt.wantLog {
				t.Errorf("logFunc() called = %v, want %v", logFuncCalled, tt.wantLog)
			}
			if tt.args.c.closeCalled {
				t.Errorf("Close() called = %v, want %v", tt.args.c.closeCalled, true)
			}
		})
	}
}

func TestNewAdapter(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer MustClose(db, nil)

	logFunc := func(v ...any) { fmt.Println(v...) }
	txo := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	type args struct {
		db      *sql.DB
		options []Option
	}
	tests := []struct {
		name string
		args args
		want *Adapter
	}{
		{
			name: "Test without options",
			args: args{
				db: db,
				options: []Option{
					WithLog(logFunc),
				},
			},
			want: &Adapter{db: db},
		},
		{
			name: "Test with txOptions",
			args: args{
				db: db,
				options: []Option{
					WithLog(logFunc),
					WithTxOptions(txo),
				},
			},
			want: &Adapter{db: db, txOptions: txo},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAdapter(tt.args.db, tt.args.options...)
			if got.db != tt.want.db {
				t.Errorf("NewAdapter() = %v, want %v", got, tt.want)
			}
		})
	}
}
