package reader

import (
	"bufio"
	"io"
	"reflect"
	"testing"
)

func TestNewSQLReader(t *testing.T) {
	type args struct {
		source io.Reader
	}
	tests := []struct {
		name string
		args args
		want *SQLReader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSQLReader(tt.args.source); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSQLReader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLReader_Next(t *testing.T) {
	type fields struct {
		source *bufio.Reader
	}
	tests := []struct {
		name      string
		fields    fields
		wantQuery string
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SQLReader{
				source: tt.fields.source,
			}
			gotQuery, err := r.Next()
			if (err != nil) != tt.wantErr {
				t.Errorf("Next() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotQuery != tt.wantQuery {
				t.Errorf("Next() gotQuery = %v, want %v", gotQuery, tt.wantQuery)
			}
		})
	}
}
