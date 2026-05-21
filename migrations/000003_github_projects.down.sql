DROP INDEX IF EXISTS idx_projects_user;
ALTER TABLE projects DROP COLUMN IF EXISTS github_branch;
ALTER TABLE projects DROP COLUMN IF EXISTS github_url;
ALTER TABLE projects DROP COLUMN IF EXISTS user_id;
