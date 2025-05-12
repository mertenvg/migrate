package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/mertenvg/migrate/pkg/reader"
	"github.com/mertenvg/migrate/pkg/statements"
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
	INSERT INTO "migrations" (name, rollback)
	VALUES ($1, $2);
`

const migrations = `
	SELECT "name" FROM "migrations" ORDER BY "name";
`

const rollbackWithName = `
	SELECT "rollback" FROM "migrations" WHERE name = $1
`

const removeWithName = `
	DELETE FROM "migrations" WHERE name = $1;
`

type LogFunc func(v ...any)

type Option func(*Adapter)

type Adapter struct {
	db        *sql.DB
	log       LogFunc
	txOptions *sql.TxOptions
	tx        *sql.Tx
	stmts     statements.Statements
}

func MustClose(c io.Closer, log LogFunc) {
	err := c.Close()
	if err != nil && log != nil {
		log("failed to close:", err)
	}
}

func NewAdapter(db *sql.DB, options ...Option) *Adapter {
	a := &Adapter{
		db:  db,
		log: func(v ...any) {},
	}
	for _, option := range options {
		option(a)
	}
	return a
}

func (a *Adapter) Setup() error {
	_, err := a.db.Exec(migrationStore)
	if err != nil {
		return fmt.Errorf("postgres.Adapter Setup failed: %w", err)
	}
	a.stmts, err = statements.Prepare(a.db, add, migrations, rollbackWithName, removeWithName)
	if err != nil {
		return fmt.Errorf("postgres.Adapter Setup failed: %w", err)
	}
	return nil
}

func (a *Adapter) List() ([]string, error) {
	rows, err := a.stmts.Get(migrations).Query()
	if err != nil {
		return nil, fmt.Errorf("postgres.Adapter List failed: %w", err)
	}
	defer MustClose(rows, a.log)

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			MustClose(rows, a.log)
			return nil, fmt.Errorf("postgres.Adapter List failed: %w", err)
		}
		names = append(names, name)
	}

	return names, nil
}

func (a *Adapter) Begin(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := a.db.BeginTx(ctx, a.txOptions)
	if err != nil {
		return fmt.Errorf("postgres.Adapter Begin failed: %w", err)
	}
	a.tx = tx
	return nil
}

func (a *Adapter) apply(r *reader.SQLReader) error {
	for {
		q, err := r.Next()
		if err != nil {
			return fmt.Errorf("failed to get query: %w", err)
		}
		if q == "" {
			break
		}
		firstWord := strings.ToUpper(strings.Split(q, " ")[0])
		if strings.HasPrefix(firstWord, "BEGIN") || strings.HasPrefix(firstWord, "COMMIT") {
			continue
		}

		_, err = a.tx.Exec(q)
		if err != nil {
			return fmt.Errorf("failed to execute query '%s': %w", q, err)
		}
	}
	return nil
}

func (a *Adapter) Up(name string, up, down io.Reader) error {
	a.log("Applying migration", name)

	err := a.apply(reader.NewSQLReader(up))
	if err != nil {
		return fmt.Errorf("postgres.Adapter Up error for migration '%s': %w", name, err)
	}

	var downData []byte

	if down != nil {
		downData, err = io.ReadAll(down)
		if err != nil {
			return fmt.Errorf("postgres.Adapter Up failed to read down file for migration '%s': %w", name, err)
		}
	}

	if _, err = a.tx.Stmt(a.stmts.Get(add)).Exec(name, string(downData)); err != nil {
		return fmt.Errorf("postgres.Adapter Up failed to register migration '%s': %w", name, err)
	}

	return nil
}

func (a *Adapter) Down(name string) error {
	a.log("Taking down migration", name)

	var rollback string
	err := a.tx.Stmt(a.stmts.Get(rollbackWithName)).QueryRow(name).Scan(&rollback)
	if err != nil {
		return fmt.Errorf("postgres.Adapter Down failed to get rollback sql: %w", err)
	}

	if rollback == "" {
		a.log("migration", name, "cannot be taken down because it does not have a rollback")
		return nil
	}

	err = a.apply(reader.NewSQLReader(bytes.NewBufferString(rollback)))
	if err != nil {
		return fmt.Errorf("postgres.Adapter Down error for migration '%s': %w", name, err)
	}

	if _, err = a.tx.Stmt(a.stmts.Get(removeWithName)).Exec(name); err != nil {
		return fmt.Errorf("postgres.Adapter Down failed to remove migration '%s': %w", name, err)
	}

	return nil
}

func (a *Adapter) Commit() error {
	if a.tx == nil {
		return fmt.Errorf("postgres.Adapter Commit failed: no transaction to commit")
	}
	err := a.tx.Commit()
	if err != nil {
		return fmt.Errorf("postgres.Adapter Commit failed: %w", err)
	}
	a.tx = nil
	return nil
}

func (a *Adapter) Rollback() error {
	if a.tx == nil {
		return fmt.Errorf("postgres.Adapter Rollback failed: no transaction to commit")
	}
	err := a.tx.Rollback()
	if err != nil {
		return fmt.Errorf("postgres.Adapter Rollback failed: %w", err)
	}
	a.tx = nil
	return nil
}
