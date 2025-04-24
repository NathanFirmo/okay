package runner

import (
	"database/sql"
	"fmt"

	"github.com/NathanFirmo/okay/internal/db"
	"github.com/NathanFirmo/okay/internal/parser"
)

type Runner struct {
	suite     *parser.Suite
	suiteName string
	captures  map[string]string
	databases map[string]*sql.DB
}

// ANSI color codes
const (
	greenBg = "\033[42;30;1m" // Green background, black text, bold
	redBg   = "\033[41;30;1m" // Red background, black text, bold
	reset   = "\033[0m"       // Reset colors
	dim     = "\033[2m"       // Dim text for secondary info
)

func NewRunner(suite *parser.Suite, suiteName string) (*Runner, error) {
	r := &Runner{
		suite:     suite,
		suiteName: suiteName,
		captures:  make(map[string]string),
		databases: make(map[string]*sql.DB),
	}

	for _, dbCfg := range suite.Settings.Databases {
		cfg := db.DatabaseConfig{
			Type:     dbCfg.Type,
			Host:     dbCfg.Host,
			Port:     dbCfg.Port,
			User:     dbCfg.User,
			Password: dbCfg.Password,
			DBName:   dbCfg.DBName,
		}
		conn, err := db.NewConnection(cfg)
		if err != nil {
			return nil, fmt.Errorf("error connecting to database %s: %v", dbCfg.Alias, err)
		}
		r.databases[dbCfg.Alias] = conn
	}

	return r, nil
}

func (r *Runner) Run() error {
	fmt.Printf("Test Suite: %s\n\n", r.suiteName)
	return r.execTests()
}
