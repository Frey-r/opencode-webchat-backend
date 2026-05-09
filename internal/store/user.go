package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username`,
		username, passwordHash,
	).Scan(&user.ID, &user.Username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}