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

func (m *MockMigration) Name() string {
	return m.name
}

func (m *MockMigration) Up() io.Reader {
	return m.up
}

func (m *MockMigration) Down() io.Reader {
	return m.down
}

func (m *MockMigration) Close() {
	// do nothing
}

type MockProvider struct {
	nextErr error
	pos     int
	names   []string
}

func (m *MockProvider) Next() (Migration, error) {
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
	return migration, m.nextErr
}

type MockAdapter struct {
	setupErr    error
	listErr     error
	beginErr    error
	upErr       error
	downErr     error
	commitErr   error
	rollbackErr error
	applied     []string
	up          []string
	down        []string
}

func (m *MockAdapter) Setup() error {
	if m.applied == nil {
		m.applied = make([]string, 0)
	}
	return m.setupErr
}

func (m *MockAdapter) List() ([]string, error) {
	return m.applied, m.listErr
}

func (m *MockAdapter) Begin(ctx context.Context) error {
	m.up = make([]string, 0)
	m.down = make([]string, 0)
	return m.beginErr
}

func (m *MockAdapter) Up(name string, up, down io.Reader) error {
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
	return m.upErr
}

func (m *MockAdapter) Down(name string) error {
	m.down = append(m.down, name)
	return m.downErr
}

func (m *MockAdapter) Commit() error {
	for _, name := range m.down {
		rmi := slices.Index(m.applied, name)
		applied := append(m.applied[0:rmi], m.applied[rmi+1:]...)
		m.applied = applied
	}
	for _, name := range m.up {
		m.applied = append(m.applied, name)
	}
	return m.commitErr
}

func (m *MockAdapter) Rollback() error {
	return m.rollbackErr
}

func TestMigrate_Migrate(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		m       *Migrate
		args    args
		wantErr bool
	}{
		{
			name: "migrate with nothing",
			m:    New(),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate without provider",
			m:    New(WithAdapter(&MockAdapter{})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate without adapter",
			m:    New(WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate with no migrations",
			m:    New(WithAdapter(&MockAdapter{}), WithProvider(&MockProvider{})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
		{
			name: "migrate with setup error",
			m:    New(WithAdapter(&MockAdapter{setupErr: fmt.Errorf("fail setup")}), WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate with provider list error",
			m:    New(WithAdapter(&MockAdapter{}), WithProvider(&MockProvider{names: []string{"aaa", "bbb", "ccc", "ddd"}, nextErr: fmt.Errorf("next error")})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate with adapter list error",
			m:    New(WithAdapter(&MockAdapter{listErr: fmt.Errorf("fail list")}), WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate with begin error",
			m:    New(WithAdapter(&MockAdapter{beginErr: fmt.Errorf("fail begin")}), WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate with up error",
			m:    New(WithAdapter(&MockAdapter{upErr: fmt.Errorf("fail up")}), WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate with down error",
			m:    New(WithAdapter(&MockAdapter{applied: []string{"bbb"}, downErr: fmt.Errorf("fail down")}), WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate with commit error",
			m:    New(WithAdapter(&MockAdapter{commitErr: fmt.Errorf("fail commit")}), WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "migrate with rollback error",
			m:    New(WithAdapter(&MockAdapter{rollbackErr: fmt.Errorf("fail rollback")}), WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
		{
			name: "migrate with previous migrations applied",
			m:    New(WithAdapter(&MockAdapter{applied: []string{"aaa"}}), WithProvider(&MockProvider{names: []string{"aaa"}})),
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.Migrate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Migrate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		opts []Option
	}
	a := &MockAdapter{}
	p := &MockProvider{}
	tests := []struct {
		name string
		args args
		want *Migrate
	}{
		{
			name: "test without options",
			args: args{},
			want: &Migrate{},
		},
		{
			name: "test with adapter",
			args: args{
				opts: []Option{
					WithAdapter(a),
				},
			},
			want: &Migrate{
				a: a,
			},
		},
		{
			name: "test with provider",
			args: args{
				opts: []Option{
					WithProvider(p),
				},
			},
			want: &Migrate{
				p: p,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
