package migrate

import (
	"context"
	"fmt"
	"slices"
)

type Migrate struct {
	a Adapter
	p Provider
}

type Option func(m *Migrate)

func New(opts ...Option) *Migrate {
	m := &Migrate{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Migrate) Migrate(ctx context.Context) error {
	if m.a == nil {
		return fmt.Errorf("no adapter provided")
	}
	if m.p == nil {
		return fmt.Errorf("no provider provided")
	}

	// make sure adapter is initiated
	if err := m.a.Setup(); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// get list of migration files from provider
	var names []string
	migrations := map[string]Migration{}
	for {
		migration, err := m.p.Next()
		if migration == nil {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to get migrations: %w", err)
		}
		names = append(names, migration.Name())
		migrations[migration.Name()] = migration
	}

	// get list of applied migrations
	applied, err := m.a.List()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// start the transaction
	if err := m.a.Begin(ctx); err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	// take down migration no longer available
	for _, name := range applied {
		_, ok := migrations[name]
		if ok {
			// migration files are still there, leave it alone
			continue
		}
		err := m.a.Down(name)
		if err != nil {
			return fmt.Errorf("failed to take down migration '%v': %w, %w", name, err, m.a.Rollback())
		}
	}

	// apply new migrations
	for _, name := range names {
		migration, ok := migrations[name]
		if !ok || migration == nil {
			// if this is missing there's something wrong
			return fmt.Errorf("failed to find migration '%v': %w, %w", name, err, m.a.Rollback())
		}
		if slices.Contains(applied, migration.Name()) {
			// this migration is already applied, we can skip it
			continue
		}
		err := m.a.Up(migration.Name(), migration.Up(), migration.Down())
		if err != nil {
			return fmt.Errorf("failed to apply migration '%v': %w, %w", name, err, m.a.Rollback())
		}
		migration.Close()
	}

	// commit the changes
	if err := m.a.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}
