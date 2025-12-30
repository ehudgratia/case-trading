package repository

import (
	"case-trading/app/helper/auth"
	"case-trading/app/helper/hash"
	"case-trading/app/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *Service) Register(ctx context.Context, input models.RegisterRequest) (*models.Users, error) {
	if input.Username == "" || input.Email == "" || input.Password == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	hashedPassword, err := hash.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user := models.Users{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
	}

	if err := s.DB.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) Login(ctx context.Context, input models.LoginRequest) (*models.LoginResponse, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	var user models.Users
	if err := s.DB.WithContext(ctx).
		Where("email = ?", input.Email).
		First(&user).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if !hash.CheckPassword(user.Password, input.Password) {
		return nil, errors.New("invalid email or password")
	}

	token, expiredAt, err := auth.CreateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	verifyToken := models.VerifyToken{
		Token:     token,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: time.Now(),
		ExpiredAt: expiredAt,
	}

	if err := s.DB.WithContext(ctx).Create(&verifyToken).Error; err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Success:   true,
		Token:     token,
		ExpiredAt: expiredAt,
	}, nil
}
