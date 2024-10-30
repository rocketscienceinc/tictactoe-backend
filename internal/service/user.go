package service

import (
	"context"
	"fmt"

	"github.com/rocketscienceinc/tittactoe-backend/internal/entity"
)

type UserService interface {
	SaveUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}

type userRepo interface {
	Save(ctx context.Context, user *entity.User) error
	Find(ctx context.Context, email string) (*entity.User, error)
}

type userService struct {
	userRepo userRepo
}

func NewUserService(userRepo userRepo) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (that *userService) SaveUser(ctx context.Context, user *entity.User) error {
	if err := that.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("could not save user: %w", err)
	}
	return nil
}

func (that *userService) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := that.userRepo.Find(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("could not get user by email: %w", err)
	}

	return user, nil
}
