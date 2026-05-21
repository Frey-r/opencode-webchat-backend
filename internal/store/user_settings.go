package store

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

type UserSettings struct {
	UserID   int
	Settings map[string]string
}

func (s *Store) GetUserSettings(ctx context.Context, userID int) (map[string]string, error) {
	var settingsJSON []byte
	err := s.pool.QueryRow(ctx,
		`SELECT settings FROM user_settings WHERE user_id = $1`,
		userID,
	).Scan(&settingsJSON)

	if err != nil {
		if err == pgx.ErrNoRows {
			return make(map[string]string), nil
		}
		return nil, err
	}

	var settings map[string]string
	if err := json.Unmarshal(settingsJSON, &settings); err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *Store) SaveUserSettings(ctx context.Context, userID int, settings map[string]string) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO user_settings (user_id, settings) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO UPDATE SET settings = EXCLUDED.settings, updated_at = CURRENT_TIMESTAMP`,
		userID, settingsJSON,
	)
	return err
}
