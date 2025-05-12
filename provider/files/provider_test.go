package files

import (
	"io"
	"reflect"
	"testing"

	"github.com/mertenvg/migrate"
)

func TestMigration_Close(t *testing.T) {
	type fields struct {
		name     string
		upPath   string
		downPath string
		close    []io.Closer
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				name:     tt.fields.name,
				upPath:   tt.fields.upPath,
				downPath: tt.fields.downPath,
				close:    tt.fields.close,
			}
			m.Close()
		})
	}
}

func TestMigration_Down(t *testing.T) {
	type fields struct {
		name     string
		upPath   string
		downPath string
		close    []io.Closer
	}
	tests := []struct {
		name   string
		fields fields
		want   io.Reader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				name:     tt.fields.name,
				upPath:   tt.fields.upPath,
				downPath: tt.fields.downPath,
				close:    tt.fields.close,
			}
			if got := m.Down(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Down() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMigration_Name(t *testing.T) {
	type fields struct {
		name     string
		upPath   string
		downPath string
		close    []io.Closer
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				name:     tt.fields.name,
				upPath:   tt.fields.upPath,
				downPath: tt.fields.downPath,
				close:    tt.fields.close,
			}
			if got := m.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMigration_Up(t *testing.T) {
	type fields struct {
		name     string
		upPath   string
		downPath string
		close    []io.Closer
	}
	tests := []struct {
		name   string
		fields fields
		want   io.Reader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				name:     tt.fields.name,
				upPath:   tt.fields.upPath,
				downPath: tt.fields.downPath,
				close:    tt.fields.close,
			}
			if got := m.Up(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Up() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewProvider(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want *Provider
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProvider(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_Next(t *testing.T) {
	type fields struct {
		position   int
		names      []string
		migrations map[string]*Migration
	}
	tests := []struct {
		name    string
		fields  fields
		want    migrate.Migration
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				position:   tt.fields.position,
				names:      tt.fields.names,
				migrations: tt.fields.migrations,
			}
			got, err := p.Next()
			if (err != nil) != tt.wantErr {
				t.Errorf("Next() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Next() got = %v, want %v", got, tt.want)
			}
		})
	}
}
