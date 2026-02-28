-- 000001_init.up.sql
-- Initial schema for Mini Airflow distributed task scheduler.
-- Requires PostgreSQL 13+ (uuid-ossp extension for uuid_generate_v4()).

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- workflows: named, schedulable DAGs of tasks.
CREATE TABLE workflows (
    id            UUID        NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    name          TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    schedule_cron TEXT        NOT NULL DEFAULT '',
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workflows_is_active  ON workflows (is_active);
CREATE INDEX idx_workflows_created_at ON workflows (created_at);

-- tasks: individual units of work belonging to a workflow.
CREATE TABLE tasks (
    id                   UUID        NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    workflow_id          UUID        NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    name                 TEXT        NOT NULL,
    command              TEXT        NOT NULL DEFAULT '',
    retry_count          INT         NOT NULL DEFAULT 0,
    retry_delay_seconds  INT         NOT NULL DEFAULT 0,
    timeout_seconds      INT         NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_workflow_id ON tasks (workflow_id);
CREATE INDEX idx_tasks_created_at  ON tasks (created_at);

-- task_dependencies: upstream/downstream relationships between tasks.
CREATE TABLE task_dependencies (
    id                UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    task_id           UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    depends_on_task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    UNIQUE (task_id, depends_on_task_id)
);

CREATE INDEX idx_task_dependencies_task_id            ON task_dependencies (task_id);
CREATE INDEX idx_task_dependencies_depends_on_task_id ON task_dependencies (depends_on_task_id);

-- workflow_runs: execution instances of a workflow.
CREATE TABLE workflow_runs (
    id          UUID        NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    workflow_id UUID        NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    status      TEXT        NOT NULL DEFAULT 'pending',
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX idx_workflow_runs_workflow_id ON workflow_runs (workflow_id);
CREATE INDEX idx_workflow_runs_status      ON workflow_runs (status);
CREATE INDEX idx_workflow_runs_started_at  ON workflow_runs (started_at);

-- task_runs: individual execution attempts of a task within a workflow run.
CREATE TABLE task_runs (
    id              UUID        NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    workflow_run_id UUID        NOT NULL REFERENCES workflow_runs (id) ON DELETE CASCADE,
    task_id         UUID        NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    status          TEXT        NOT NULL DEFAULT 'pending',
    attempt         INT         NOT NULL DEFAULT 1,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ,
    logs            TEXT        NOT NULL DEFAULT ''
);

CREATE INDEX idx_task_runs_workflow_run_id ON task_runs (workflow_run_id);
CREATE INDEX idx_task_runs_task_id         ON task_runs (task_id);
CREATE INDEX idx_task_runs_status          ON task_runs (status);
CREATE INDEX idx_task_runs_started_at      ON task_runs (started_at);

-- workers: nodes that poll for and execute tasks.
CREATE TABLE workers (
    id             UUID        NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    hostname       TEXT        NOT NULL,
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status         TEXT        NOT NULL DEFAULT 'active'
);

CREATE INDEX idx_workers_status          ON workers (status);
CREATE INDEX idx_workers_last_heartbeat  ON workers (last_heartbeat);
