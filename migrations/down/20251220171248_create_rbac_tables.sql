-- migration down

-- Remove role_id column from users table
ALTER TABLE `users` DROP COLUMN `role_id`;

-- Drop tables in reverse order (respecting dependencies)
DROP TABLE IF EXISTS `role_permissions`;
DROP TABLE IF EXISTS `permissions`;
DROP TABLE IF EXISTS `roles`;
