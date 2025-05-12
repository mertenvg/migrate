package reader

import (
	"bufio"
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestNewSQLReader(t *testing.T) {
	src := bytes.NewBufferString("")
	want := &SQLReader{
		source: bufio.NewReader(src),
	}
	if got := NewSQLReader(src); !reflect.DeepEqual(got, want) {
		t.Errorf("NewSQLReader() = %v, want %v", got, want)
	}
}

func validateNext(t *testing.T, r *SQLReader, want string, wantErr bool) {
	q, err := r.Next()
	if (err != nil) != wantErr {
		t.Errorf("Next() unexpected error = %v, want nil", err)
	}
	if q != want {
		t.Errorf("Next() got '%v', want '%v'", q, want)
	}
}

func TestSQLReader_Next(t *testing.T) {
	src := bytes.NewBufferString(`Query 1; Query 2; Insert "\";\"", "\\" into "table;"; done `)
	r := NewSQLReader(src)

	validateNext(t, r, "Query 1", false)
	validateNext(t, r, "Query 2", false)
	validateNext(t, r, `Insert "\";\"", "\\" into "table;"`, false)
	validateNext(t, r, "done", false)
	validateNext(t, r, "", false)
}

func TestSQLReader_Next_WithFail(t *testing.T) {
	src := bytes.NewBufferString(`Query 1; Query 2`)
	r := NewSQLReader(src)

	FailNext = errors.New("fail")

	validateNext(t, r, "", true)
	validateNext(t, r, "Query 1", false)
	validateNext(t, r, "Query 2", false)
}

type FailReader struct{}

func (r FailReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("fail")
}

func TestSQLReader_Next_WithReadFail(t *testing.T) {
	r := NewSQLReader(&FailReader{})

	validateNext(t, r, "", true)
	validateNext(t, r, "", true)
}
