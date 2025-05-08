package migrate

import "io"

// Adapter applies the necessary migrations to the database
type Adapter interface {
	// Setup must set up the migration store to record which migrations have been commited to the db
	Setup() error
	// List the applied migrations
	List() ([]string, error)
	// Begin a transaction
	Begin() error
	// Apply the migration
	Apply(name string, up, down io.Reader) error
	// Rollback a migration that was previously applied (if it has a rollback)
	Rollback(name string) error
	// Commit the transaction and store the migration name in the migration store along with an optional down/rollback migration
	Commit() error
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
}
