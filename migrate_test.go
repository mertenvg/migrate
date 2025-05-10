package migrate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"slices"
	"testing"
)

type MockMigration struct {
	name string
	up   *bytes.Buffer
	down *bytes.Buffer
}

func (m MockMigration) Name() string {
	return m.name
}

func (m MockMigration) Up() io.Reader {
	return m.up
}

func (m MockMigration) Down() io.Reader {
	return m.down
}

func (m MockMigration) Close() {
	// do nothing
}

type MockProvider struct {
	pos   int
	names []string
}

func (m MockProvider) Next() (Migration, error) {
	if m.pos >= len(m.names) {
		return nil, nil
	}
	name := m.names[m.pos]
	migration := &MockMigration{
		name: name,
		up:   bytes.NewBufferString("up " + name),
		down: bytes.NewBufferString("down " + name),
	}
	m.pos++
	return migration, nil
}

type MockAdapter struct {
	applied []string
	up      []string
	down    []string
}

func (m MockAdapter) Setup() error {
	if m.applied == nil {
		m.applied = make([]string, 0)
	}
}

func (m MockAdapter) List() ([]string, error) {
	return m.applied, nil
}

func (m MockAdapter) Begin(ctx context.Context) error {
	m.up = make([]string, 0)
	m.down = make([]string, 0)
	return nil
}

func (m MockAdapter) Up(name string, up, down io.Reader) error {
	upStr, err := io.ReadAll(up)
	if err != nil {
		return err
	}
	if string(upStr) != fmt.Sprintf("up %s", name) {
		return fmt.Errorf("up %s not valid", name)
	}
	if down != nil {
		downStr, err := io.ReadAll(down)
		if err != nil {
			return err
		}
		if string(downStr) != fmt.Sprintf("down %s", name) {
			return fmt.Errorf("down %s not valid", name)
		}
	}
	m.up = append(m.up, name)
	return nil
}

func (m MockAdapter) Down(name string) error {
	m.down = append(m.down, name)
	return nil
}

func (m MockAdapter) Commit() error {
	for _, name := range m.down {
		rmi := slices.Index(m.applied, name)
		applied := append(m.applied[0:rmi], m.applied[rmi+1:]...)
		m.applied = applied
	}
	for _, name := range m.up {
		m.applied = append(m.applied, name)
	}
	return nil
}

func (m MockAdapter) Rollback() error {
	// do nothing
}

func TestMigrate_Migrate(t *testing.T) {
	type fields struct {
		a Adapter
		p Provider
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migrate{
				a: tt.fields.a,
				p: tt.fields.p,
			}
			if err := m.Migrate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Migrate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want *Migrate
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
