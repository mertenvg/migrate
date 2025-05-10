package statements

import (
	"database/sql"
	"fmt"
)

type Statements map[string]*sql.Stmt

func (stmts Statements) Get(q string) *sql.Stmt {
	return stmts[q]
}

func Prepare(db *sql.DB, queries ...string) (Statements, error) {
	stmts := make(Statements)
	for _, q := range queries {
		stmt, err := db.Prepare(q)
		if err != nil {
			return nil, fmt.Errorf("unable to prepare query '%s' with error: %w", q, err)
		}
		stmts[q] = stmt
	}
	return stmts, nil
}
