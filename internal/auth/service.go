package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"nms_lte/internal/id"
	"nms_lte/internal/model"
)

var (
	ErrInvalidCredentials = errors.New("invalid email, username or password")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserAlreadyExists  = errors.New("username already exists")
)

type ValidationErrors map[string]string

func (e ValidationErrors) Error() string {
	return fmt.Sprintf("validation errors: %v", map[string]string(e))
}

type Store interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Register(ctx context.Context, email, username, password string) (*model.User, error) {
	errs := validateRegisterInput(email, username, password)
	if len(errs) > 0 {
		return nil, errs
	}

	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	_, err := s.store.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}

	_, err = s.store.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := model.User{
		ID:           id.New("user"),
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) Login(ctx context.Context, email, username, password string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	if password == "" || (email == "" && username == "") {
		return "", ErrInvalidCredentials
	}

	var (
		user model.User
		err  error
	)

	if email != "" {
		user, err = s.store.GetUserByEmail(ctx, email)
	} else {
		user, err = s.store.GetUserByUsername(ctx, username)
	}

	if err != nil {
		return "", ErrInvalidCredentials
	}

	if ok := CheckPasswordHash(password, user.PasswordHash); !ok {
		return "", ErrInvalidCredentials
	}

	token, err := GenerateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func validateRegisterInput(email, username, password string) ValidationErrors {
	errs := ValidationErrors{}

	email = strings.TrimSpace(email)
	username = strings.TrimSpace(username)

	if email == "" {
		errs["email"] = "email is required"
	} else if !strings.Contains(email, "@") {
		errs["email"] = "invalid email"
	}

	if username == "" {
		errs["username"] = "username is required"
	}

	if password == "" {
		errs["password"] = "password is required"
	} else if len(password) < 5 {
		errs["password"] = "password length should be at least 5"
	}

	return errs
}