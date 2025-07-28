package repository

import (
	"Grampus/internal/ledger"
	"Grampus/pkg"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// 与具体的数据库实现无关的,可以通用的公用函数

func withTransaction(db *sqlx.DB, handler func(*sqlx.Tx) error) (err error) {
	tx, err := db.BeginTxx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	if err != nil {
		zap.L().Error("Begin transaction in sqlite3 failed!", zap.Error(err))
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			// rollback if panic
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	err = handler(tx)
	if err == nil {
		tx.Commit()
	}

	return err
}

func queryWithIterator[T any](db sqlx.Ext, querySql string, args ...any) (ledger.RepoIterator[T], error) {
	rows, err := db.Queryx(querySql, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			it := pkg.NewEmptyItertor[T]()
			return &it, nil
		}
		return nil, err
	}

	return pkg.NewRowsIterator[T](rows), nil
}

func queryOne[T any](db sqlx.Ext, querySql string, args ...interface{}) (*T, error) {
	var res T
	err := db.QueryRowx(querySql, args...).StructScan(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func insertOne[T any](db sqlx.Ext, sql string, data *T) error {
	rows, err := sqlx.NamedQuery(db, sql, data)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(data); err != nil {
			return err
		}
	}
	return nil
}
