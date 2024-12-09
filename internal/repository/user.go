package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type userRepository struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) UserRepository {
	return &userRepository{
		conn: conn,
	}
}

func (that *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (email) VALUES (?)`

	_, err := that.conn.ExecContext(ctx, query, user.Email)
	if err != nil {
		return fmt.Errorf("can't save user: %w", err)
	}

	return nil
}

func (that *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT email FROM users WHERE email = ?`

	var user entity.User

	err := that.conn.QueryRowContext(ctx, query, email).Scan(&user.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("can't find user: %w", err)
	}

	return &user, nil
}
