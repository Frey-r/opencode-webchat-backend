package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Project struct {
	ID           int       `json:"id"`
	UserID       *int      `json:"user_id"`
	Name         string    `json:"name"`
	RootPath     string    `json:"root_path"`
	GitHubURL    string    `json:"github_url,omitempty"`
	GitHubBranch string    `json:"github_branch,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsedAt   time.Time `json:"last_used_at"`
}

func (s *Store) ListProjects(ctx context.Context, userID int) ([]Project, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name, root_path, 
		 COALESCE(github_url, ''), COALESCE(github_branch, ''),
		 created_at, last_used_at 
		 FROM projects 
		 WHERE user_id IS NULL OR user_id = $1 
		 ORDER BY last_used_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.RootPath,
			&p.GitHubURL, &p.GitHubBranch, &p.CreatedAt, &p.LastUsedAt); err != nil {
			return nil, err
		}
		p.RootPath = expandTilde(p.RootPath)
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *Store) CreateProject(ctx context.Context, userID *int, name, rootPath, githubURL, githubBranch string) (*Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`INSERT INTO projects (user_id, name, root_path, github_url, github_branch)
		 VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))
		 RETURNING id, user_id, name, root_path, 
		 COALESCE(github_url, ''), COALESCE(github_branch, ''),
		 created_at, last_used_at`,
		userID, name, rootPath, githubURL, githubBranch,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.RootPath,
		&p.GitHubURL, &p.GitHubBranch, &p.CreatedAt, &p.LastUsedAt)
	if err != nil {
		return nil, err
	}
	p.RootPath = expandTilde(p.RootPath)
	return &p, nil
}

func (s *Store) GetProject(ctx context.Context, id int) (*Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, name, root_path, 
		 COALESCE(github_url, ''), COALESCE(github_branch, ''),
		 created_at, last_used_at 
		 FROM projects WHERE id = $1`, id,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.RootPath,
		&p.GitHubURL, &p.GitHubBranch, &p.CreatedAt, &p.LastUsedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p.RootPath = expandTilde(p.RootPath)
	return &p, nil
}

func (s *Store) DeleteProject(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	return err
}

func (s *Store) UpdateProjectLastUsed(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE projects SET last_used_at = NOW() WHERE id = $1`, id)
	return err
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~" {
			return home
		}
		if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
