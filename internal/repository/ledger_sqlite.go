package repository

import (
	"Grampus/internal/config"
	"Grampus/internal/ledger"
	"Grampus/pkg"
	"embed"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const queryLedgerByIdSql = `SELECT id, name, defalut_curr, created_at FROM ledger where id = ?`
const queryLedgerByNameSql = `SELECT id, name, defalut_curr, created_at FROM ledger where name = ?`
const queryAccoutByIdSql = `SELECT id, ledger_id, parent_id, goods_id, level, name, path, type, current_num, current_denom, created_at FROM account WHERE id = ?`
const queryAccoutByLedgerIdSql = `SELECT id, ledger_id, parent_id, goods_id, level, name, path, type, current_num, current_denom, created_at FROM account WHERE ledger_id = ?`
const insertAccountSql = `INSERT INTO account ( id, ledger_id, parent_id, goods_id, level, name, type, current_num, current_denom, path )
		VALUES (:id, :ledger_id, :parent_id, :goods_id, :level, :name, :type, :current_num, :current_denom, :path)
		RETURNING created_at`

type SqliteRepo struct {
	db     *sqlx.DB
	closed bool
}

func (r *SqliteRepo) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	r.db.Close()
	return nil
}

func NewSqliteRepo(config *config.DatabaseConfig) (ledger.Repo, error) {
	dir := filepath.Dir(config.DSN)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			zap.L().Error("Failed to create directory for SQLite database", zap.Error(err))
			return nil, err
		}
	}

	db, err := sqlx.Connect(config.Driver, config.DSN)
	if err != nil {
		zap.L().Error("Failed to connect to SQLite database", zap.Error(err))
		return nil, err
	}

	if err := db.Ping(); err != nil {
		zap.L().Error("Failed to ping SQLite database", zap.Error(err))
		return nil, err
	}
	zap.L().Info("Connected to SQLite database successfully", zap.String("DSN", config.DSN))
	// TODO 后续这块配置也通过DataBaseconfig配置
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	repo := &SqliteRepo{
		db:     db,
		closed: false,
	}
	if err := repo.init(); err != nil {
		repo.Close()
		zap.L().Error("Init sqlite database failed", zap.Error(err))
		return nil, err
	}
	return repo, nil
}

//go:embed migrations/sqlite/*.sql
var sqliteMigrationsFs embed.FS

func (r *SqliteRepo) init() error {
	sourceDriver, err := iofs.New(sqliteMigrationsFs, "migrations/sqlite")
	if err != nil {
		return err
	}

	dbDriver, err := sqlite3.WithInstance(r.db.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	if err == migrate.ErrNoChange {
		zap.L().Info("Init ledger database: no changes to apply")
	} else {
		zap.L().Info("Init ledger database success!")
	}

	return nil
}

func (r *SqliteRepo) withForeignKeys(handler func(*sqlx.DB) error) error {
	_, err := r.db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		zap.L().Error("PRAGMA foreign keys failed", zap.Error(err))
		return errors.New("PRAGMA foreign keys failed")
	}

	return handler(r.db)
}

func (r *SqliteRepo) CreateLedger(ledger *ledger.Ledger) error {
	const insertSql = `INSERT INTO ledger (id, name, defalut_curr) VALUES (?, ?, ?) RETURNING created_at`
	ledger.ID = uuid.New().String()

	return r.withForeignKeys(func(db *sqlx.DB) error {
		return withTransaction(db, func(tx *sqlx.Tx) error {
			err := tx.QueryRowx(insertSql, ledger.ID, ledger.Name, ledger.DefalutCurr).StructScan(ledger)
			if err != nil {
				return err
			}
			return nil
		})
	})
}

func (r *SqliteRepo) GetAllLedger() (ledger.RepoIterator[ledger.Ledger], error) {
	const querySql = "SELECT id, name, defalut_curr, created_at FROM ledger;"

	it, err := queryWithIterator[ledger.Ledger](r.db, querySql)
	return it, err
}

func (r *SqliteRepo) GetLedgerById(id string) (*ledger.Ledger, error) {
	// const querySql = "SELECT id, name, defalut_curr, created_at FROM ledger where id = ?"
	return queryOne[ledger.Ledger](r.db, queryLedgerByIdSql, id)
}

func (r *SqliteRepo) GetLedgerByName(name string) (*ledger.Ledger, error) {
	return queryOne[ledger.Ledger](r.db, queryLedgerByNameSql, name)
}

func (r *SqliteRepo) GetAccountById(id string) (*ledger.Account, error) {
	return queryOne[ledger.Account](r.db, queryAccoutByIdSql, id)
}

func (r *SqliteRepo) GetLedgerAccounts(ledgerId string) (ledger.RepoIterator[ledger.Account], error) {
	return queryWithIterator[ledger.Account](r.db, queryAccoutByLedgerIdSql, ledgerId)
}

func (r *SqliteRepo) CreateAccount(account *ledger.Account) error {
	if account.LedgerID == "" {
		return pkg.ErrLedgerNotExist
	}
	return r.withForeignKeys(func(db *sqlx.DB) error {
		return withTransaction(db, func(tx *sqlx.Tx) error {
			l, err := queryOne[ledger.Ledger](tx, queryLedgerByIdSql, account.LedgerID)
			if err != nil {
				zap.L().Error("Can't get ledger by id", zap.String("ledgerId", account.LedgerID), zap.Error(err))
				return pkg.ErrLedgerNotExist
			}
			if account.GoodsID == "" {
				account.GoodsID = l.DefalutCurr
			}
			account.Level = 0
			// 校验Account的
			if account.ParentID != "" {
				pAccount, err := queryOne[ledger.Account](tx, queryAccoutByIdSql, account.ParentID)
				if err != nil {
					zap.L().Error(
						"Can't get parent account by id",
						zap.String("parent_id", account.ParentID),
						zap.Error(err),
					)
					return pkg.ErrParentAccountNotExist
				}
				if pAccount.LedgerID != account.LedgerID {
					zap.L().Error(
						"Parent account and account belone different ledger",
						zap.String("parent_account_ledger", pAccount.LedgerID),
						zap.String("account_ledger", account.LedgerID),
					)
					return pkg.ErrInvalidAccount
				}
				if pAccount.Level+1 > ledger.MaxAccountLevel {
					zap.L().Error(
						"Exceed account max nest level",
						zap.Int("account_max_nest_level", ledger.MaxAccountLevel),
						zap.Int("parent_account_level", pAccount.Level),
					)
					return pkg.ErrExceedAccountNestLevel
				}
				account.Level = pAccount.Level + 1
				account.Path = pAccount.Path + "/" + account.Name
			} else {
				account.Path = account.Name
			}

			account.ID = uuid.New().String()
			// TODO 正式插入数据
			err = insertOne(tx, insertAccountSql, account)
			if err != nil {
				zap.L().Error(
					"insert account to sqlite failed",
					zap.Any("account", account),
					zap.Error(err),
				)
				return pkg.ErrDBInsertFailed
			}
			return nil
		})
	})
}
