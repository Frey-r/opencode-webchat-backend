package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Session struct {
	ID                int
	ProjectID         int
	Title             string
	OpenCodeSessionID string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	IsActive          bool
}

func (s *Store) CreateSession(ctx context.Context, projectID int, title string) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx,
		`INSERT INTO sessions (project_id, title) VALUES ($1, $2)
		 RETURNING id, project_id, title, opencode_session_id, created_at, updated_at, is_active`,
		projectID, title,
	).Scan(&session.ID, &session.ProjectID, &session.Title, &session.OpenCodeSessionID, &session.CreatedAt, &session.UpdatedAt, &session.IsActive)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *Store) GetSessionsByProject(ctx context.Context, projectID int) ([]Session, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, title, opencode_session_id, created_at, updated_at, is_active
		 FROM sessions WHERE project_id = $1 AND is_active = TRUE ORDER BY updated_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var session Session
		if err := rows.Scan(&session.ID, &session.ProjectID, &session.Title, &session.OpenCodeSessionID, &session.CreatedAt, &session.UpdatedAt, &session.IsActive); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (s *Store) GetSessionByID(ctx context.Context, id int) (*Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx,
		`SELECT id, project_id, title, opencode_session_id, created_at, updated_at, is_active
		 FROM sessions WHERE id = $1`,
		id,
	).Scan(&session.ID, &session.ProjectID, &session.Title, &session.OpenCodeSessionID, &session.CreatedAt, &session.UpdatedAt, &session.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *Store) UpdateSessionTitle(ctx context.Context, id int, title string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET title = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		title, id,
	)
	return err
}

func (s *Store) UpdateSessionUpdatedAt(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
		id,
	)
	return err
}

func (s *Store) DeleteSession(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET is_active = FALSE WHERE id = $1`,
		id,
	)
	return err
}