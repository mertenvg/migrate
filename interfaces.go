package migrate

import (
	"context"
	"io"
)

// Adapter applies the necessary migrations to the database
type Adapter interface {
	// Setup must set up the migration store to record which migrations have been commited to the db
	Setup() error
	// List the applied migrations
	List() ([]string, error)
	// Begin a transaction
	Begin(ctx context.Context) error
	// Up a migration
	Up(name string, up, down io.Reader) error
	// Down a migration that was previously applied, if it has a down/rollback saved when it was applied
	Down(name string) error
	// Commit the transaction
	Commit() error
	// Rollback the transaction
	Rollback() error
}

// Provider tells us what needs to be done
type Provider interface {
	// Next should return the next Migration or nil if there isn't one. error is reserved for actual errors.
	Next() (Migration, error)
}

// Migration describes the migration and what needs to be applied (or rolled back)
type Migration interface {
	// Name of the migration, must be unique to avoid migration conflicts
	Name() string
	// Up returns what needs to be applied as an io.Reader
	Up() io.Reader
	// Down returns what will need to be rolled back as an io.Reader
	Down() io.Reader
	// Close any open readers or files
	Close()
}
