package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/NathanFirmo/okay/internal/parser"
)

type DatabaseConfig struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewConnection(cfg DatabaseConfig) (*sql.DB, error) {
	var dsn string
	switch cfg.Type {
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
	db, err := sql.Open(cfg.Type, dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func RunSQLFile(db *sql.DB, file string) error {
	sql, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(sql))
	return err
}

func RunSeeds(r *parser.Suite, dbs map[string]*sql.DB) error {
	for _, seed := range r.BeforeEach.Seeds {
		dbConn, ok := dbs[seed.DB]
		if !ok {
			return fmt.Errorf("database alias not found: %s", seed.DB)
		}
		err := RunSQLFile(dbConn, seed.File)
		if err != nil {
			return fmt.Errorf("error running seed for %s: %v", seed.DB, err)
		}
	}

	return nil
}
