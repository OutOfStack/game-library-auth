-- +migrate Up
-- Add new `role` column as NULLABLE
ALTER TABLE users ADD COLUMN role TEXT CHECK (role IN ('user', 'publisher'));

-- Update `users.role` based on `roles.name`
UPDATE users
SET role = roles.name
FROM roles
WHERE users.role_id = roles.id;

-- Alter `role` column to be NOT NULL
ALTER TABLE users ALTER COLUMN role SET NOT NULL;

-- Drop `users.role_id` column
ALTER TABLE users DROP COLUMN role_id;

-- Drop `roles` table
DROP TABLE roles;

-- Add index on role column
CREATE INDEX idx_users_role ON users(role) INCLUDE (id);

-- +migrate Down

