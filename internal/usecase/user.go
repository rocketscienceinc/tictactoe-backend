package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

type UserUseCase interface {
	GetOrCreate(ctx context.Context, user *entity.User) (*entity.User, error)
}

type userRepo interface {
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
}

type userUseCase struct {
	repo userRepo
}

func NewUserUseCase(repo userRepo) UserUseCase {
	return &userUseCase{
		repo: repo,
	}
}

func (that *userUseCase) GetOrCreate(ctx context.Context, user *entity.User) (*entity.User, error) {
	existingUser, err := that.repo.GetByEmail(ctx, user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// user not found, create new user
			err = that.repo.Create(ctx, user)
			if err != nil {
				return nil, fmt.Errorf("could not create user: %w", err)
			}
			return user, nil
		}
		return nil, fmt.Errorf("could not get user: %w", err)
	}
	return existingUser, nil
}
