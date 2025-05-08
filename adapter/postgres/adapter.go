package postgres

import (
	"database/sql"
	"fmt"
	"io"
)

const migrationStore = `
	CREATE TABLE IF NOT EXISTS "migrations" (
		"name" VARCHAR(255) NOT NULL,
		"created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
		"rollback" TEXT NULL,
		CONSTRAINT "migrations_pkey" PRIMARY KEY ("name")
	)
`

const add = `
	INSERT INTO "migrations" (name, down)
	VALUES ($1, $2);
`

const selectNames = `
	SELECT "name" FROM "migrations" ORDER BY "name";
`

const getRollback = `
	SELECT "rollback" FROM "migrations" WHERE name = $1
`

const remove = `
	DELETE FROM "migrations" WHERE name = $1;
`

type Adapter struct {
	db *sql.DB
}

func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{
		db: db,
	}
}

func (a *Adapter) Setup() error {
	_, err := a.db.Exec(migrationStore)
	if err != nil {
		return fmt.Errorf("postgres.Adapter Setup failed: %w", err)
	}
	return nil
}

func (a *Adapter) List() ([]string, error) {
	// TODO implement me
	panic("implement me")
}

func (a *Adapter) Begin() error {
	// TODO implement me
	panic("implement me")
}

func (a *Adapter) Apply(name string, up, down io.Reader) error {
	// TODO implement me
	panic("implement me")
}

func (a *Adapter) Rollback(name string) error {
	// TODO implement me
	panic("implement me")
}

func (a *Adapter) Commit() error {
	// TODO implement me
	panic("implement me")
}
