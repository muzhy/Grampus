package repository

import (
	"Grampus/internal/config"
	"Grampus/internal/ledger"
	"errors"
)

func NewLedgerRepository(config *config.DatabaseConfig) (ledger.Repo, error) {
	switch config.Driver {
	case "sqlite3":
		return NewSqliteRepo(config)
	default:
		return nil, errors.New("unknow database type")
	}
}
