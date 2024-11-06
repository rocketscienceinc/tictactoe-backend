package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	// import the SQLite driver to register it with the database/sql package.
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	Connection *sql.DB
}

func New(path string) (*Storage, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w", err)
	}

	return &Storage{Connection: conn}, nil
}

func (that *Storage) Init(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS users (email TEXT)`

	_, err := that.Connection.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}

	return nil
}
