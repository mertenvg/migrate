package statements

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestPrepare(t *testing.T) {
	type args struct {
		db      *sql.DB
		queries []string
	}
	tests := []struct {
		name    string
		args    args
		want    Statements
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Prepare(tt.args.db, tt.args.queries...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Prepare() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatements_Get(t *testing.T) {
	type args struct {
		q string
	}
	tests := []struct {
		name  string
		stmts Statements
		args  args
		want  *sql.Stmt
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stmts.Get(tt.args.q); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
