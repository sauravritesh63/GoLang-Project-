-- 000001_init.down.sql
-- Rolls back the initial schema migration.

DROP TABLE IF EXISTS task_runs;
DROP TABLE IF EXISTS workflow_runs;
DROP TABLE IF EXISTS task_dependencies;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS workflows;
DROP TABLE IF EXISTS workers;
