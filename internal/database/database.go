package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/secure-review/internal/entity"
)

// Database wraps GORM DB - аналог DataSource в TypeORM
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection - аналог new DataSource({...}).initialize()
func NewDatabase(dsn string) (*Database, error) {
	// Configure GORM logger - аналог logging: true в TypeORM
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open connection with GORM - аналог TypeORM connection options
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true, // Improve performance
		PrepareStmt:            true, // Cache prepared statements
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool - аналог TypeORM pool options
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Database{DB: db}, nil
}

// AutoMigrate runs auto migrations - аналог synchronize: true в TypeORM
func (d *Database) AutoMigrate() error {
	return d.DB.AutoMigrate(
		&entity.User{},
		&entity.CodeReview{},
		&entity.SecurityIssue{},
	)
}

// Transaction executes a function within a database transaction
// Аналог manager.transaction() в TypeORM
func (d *Database) Transaction(fn func(tx *gorm.DB) error) error {
	return d.DB.Transaction(fn)
}

// WithPreload returns a DB instance with preloaded relations
// Аналог { relations: [...] } в TypeORM
func (d *Database) WithPreload(relations ...string) *gorm.DB {
	db := d.DB
	for _, relation := range relations {
		db = db.Preload(relation)
	}
	return db
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the underlying GORM DB instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Ping checks the database connection
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
