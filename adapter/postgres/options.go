package postgres

import "database/sql"

func WithLog(f LogFunc) Option {
	return func(a *Adapter) {
		a.log = f
	}
}

func WithTxOptions(txOptions *sql.TxOptions) Option {
	return func(a *Adapter) {
		a.txOptions = txOptions
	}
}
