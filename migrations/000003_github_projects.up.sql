-- Add user_id and GitHub fields to projects
ALTER TABLE projects ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS github_url VARCHAR(1024);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS github_branch VARCHAR(255) DEFAULT 'main';

-- Index
CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id);
