package repository

import (
	"Grampus/internal/books"
	"Grampus/internal/config"
	"os"
	"path/filepath"
	"time"

	"Grampus/internal/repository/sqlite"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Repository struct {
	BookRepository books.BookRepository
}

func NewRepository(config *config.DatabaseConfig) *Repository {
	switch config.Driver {
	case "sqlite3":
		repo, _ := createSqliteRepo(config)
		return repo
	case "postgres":
		return createPostgresRepo(config)
	default:
		zap.L().Error("Unknow database drive", zap.String("drive", config.Driver))
	}
	return nil
}

func createSqliteRepo(config *config.DatabaseConfig) (*Repository, error) {
	dsn := config.DSN
	dir := filepath.Dir(dsn)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			zap.L().Error("Failed to create directory for SQLite database", zap.Error(err))
			return nil, err
		}
	}

	db, err := sqlx.Connect(config.Driver, dsn)
	if err != nil {
		zap.L().Error("Failed to connect to SQLite database", zap.Error(err))
		return nil, err
	}
	if err := db.Ping(); err != nil {
		zap.L().Error("Failed to ping SQLite database", zap.Error(err))
		return nil, err
	}
	zap.L().Info("Connected to SQLite database successfully", zap.String("DSN", dsn))

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	sqliteBooksRepo := sqlite.NewSqliteBooksRepository(db, config)
	if err := sqliteBooksRepo.InitBooksDB(); err != nil {
		zap.L().Error("Init sqlite database failed", zap.Error(err))
		return nil, err
	}

	return &Repository{
		BookRepository: sqliteBooksRepo,
	}, nil
}

func createPostgresRepo(config *config.DatabaseConfig) *Repository {
	zap.L().Error("Postgres repository is not implemented yet")
	return nil // TODO: Implement Postgres repository initialization
}
