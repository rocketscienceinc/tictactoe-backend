package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/rocketscienceinc/tictactoe-backend/internal/apperror"
	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

type UserUseCase interface {
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
}

type userRepo interface {
	Save(ctx context.Context, user *entity.User) error
	Find(ctx context.Context, email string) (*entity.User, error)
}

type userUseCase struct {
	repo userRepo
}

func NewUserUseCase(repo userRepo) UserUseCase {
	return &userUseCase{
		repo: repo,
	}
}

func (that *userUseCase) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	user, err := that.repo.Find(ctx, user.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			if err = that.repo.Save(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to save user into storage: %s", err)
			}
			return user, nil
		}
		return nil, fmt.Errorf("failed to find user into storage: %s", err)
	}

	return user, nil
}
