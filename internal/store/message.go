package store

import (
	"context"
	"encoding/json"
	"time"
)

type Message struct {
	ID        int
	SessionID int
	Role      string
	Content   string
	ToolCalls []ToolCall
	Status    string
	CreatedAt time.Time
}

type ToolCall struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (s *Store) CreateMessage(ctx context.Context, sessionID int, role, content string, toolCalls []ToolCall, status string) (*Message, error) {
	tcJSON, _ := json.Marshal(toolCalls)
	var msg Message
	err := s.pool.QueryRow(ctx,
		`INSERT INTO messages (session_id, role, content, tool_calls, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, session_id, role, content, tool_calls, status, created_at`,
		sessionID, role, content, tcJSON, status,
	).Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &tcJSON, &msg.Status, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(tcJSON, &msg.ToolCalls)
	return &msg, nil
}

func (s *Store) GetMessagesBySession(ctx context.Context, sessionID int, limit, offset int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, role, content, tool_calls, status, created_at
		 FROM messages WHERE session_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		sessionID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var tcJSON []byte
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &tcJSON, &msg.Status, &msg.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(tcJSON, &msg.ToolCalls)
		messages = append(messages, msg)
	}
	return messages, nil
}

func (s *Store) SearchMessages(ctx context.Context, query string, limit int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, role, content, tool_calls, status, created_at
		 FROM messages
		 WHERE to_tsvector('simple', content) @@ plainto_tsquery('simple', $1)
		    OR content ILIKE '%' || $1 || '%'
		 ORDER BY created_at DESC LIMIT $2`,
		query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var tcJSON []byte
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &tcJSON, &msg.Status, &msg.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(tcJSON, &msg.ToolCalls)
		messages = append(messages, msg)
	}
	return messages, nil
}