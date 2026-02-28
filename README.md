# Mini Airflow — Distributed Task Scheduler

A lightweight, distributed task scheduler inspired by Apache Airflow, built with Go and clean architecture principles.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Mini Airflow System                          │
│                                                                     │
│   ┌───────────┐     ┌───────────────┐     ┌─────────────────────┐  │
│   │  Scheduler│────▶│  Task Queue   │────▶│  Worker Pool        │  │
│   │  (cron)   │     │  (internal)   │     │  worker-1 worker-2  │  │
│   └───────────┘     └───────────────┘     └─────────────────────┘  │
│         │                                           │               │
│         ▼                                           ▼               │
│   ┌───────────┐                           ┌─────────────────────┐  │
│   │  Workflow │                           │  Task Executor      │  │
│   │  Store    │                           │  (cmd runner)       │  │
│   └───────────┘                           └─────────────────────┘  │
│                                                     │               │
│                                                     ▼               │
│                                           ┌─────────────────────┐  │
│                                           │  Run Log / Status   │  │
│                                           │  Store              │  │
│                                           └─────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

Clean-architecture package layout:

```
┌─────────────────────────────────────────────┐
│            internal/domain/                 │  ← Phase 1 ✅ (Workflow, Task, Worker models)
│            domain/                          │  ← Phase 1 ✅ (Scheduler interfaces & entities)
└───────────────────┬─────────────────────────┘
                    │ (interfaces)
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   scheduler/   worker/    storage/           ← Phase 2 (next)
  (use-cases) (executor)  (in-memory / Redis)
        │
        ▼
    api/http                                  ← Phase 3
    (REST endpoints)
```

---

## Getting Started

```bash
# Clone the repository
git clone https://github.com/sauravritesh63/GoLang-Project-

# Build
go build ./...

# Vet (static analysis)
go vet ./...

# Run tests
go test ./...
```

---

## Domain Models (`internal/domain`)

### Status Enums

| Type           | Values                                      |
|----------------|---------------------------------------------|
| `Status`       | `pending`, `running`, `success`, `failed`   |
| `WorkerStatus` | `active`, `inactive`                        |

### Structs

#### `Workflow`
Represents a named, schedulable DAG of tasks.

| Field          | Type        | JSON key        | Description                    |
|----------------|-------------|-----------------|--------------------------------|
| `ID`           | `uuid.UUID` | `id`            | Unique workflow identifier     |
| `Name`         | `string`    | `name`          | Human-readable workflow name   |
| `Description`  | `string`    | `description`   | Optional description           |
| `ScheduleCron` | `string`    | `schedule_cron` | Cron expression for scheduling |
| `IsActive`     | `bool`      | `is_active`     | Whether the workflow is enabled|
| `CreatedAt`    | `time.Time` | `created_at`    | Creation timestamp             |

#### `Task`
A single unit of work belonging to a `Workflow`.

| Field               | Type        | JSON key               | Description                                |
|---------------------|-------------|------------------------|--------------------------------------------|
| `ID`                | `uuid.UUID` | `id`                   | Unique task identifier                     |
| `WorkflowID`        | `uuid.UUID` | `workflow_id`          | Parent workflow                            |
| `Name`              | `string`    | `name`                 | Task name                                  |
| `Command`           | `string`    | `command`              | Shell command or executable to run         |
| `RetryCount`        | `int`       | `retry_count`          | Number of retry attempts on failure        |
| `RetryDelaySeconds` | `int`       | `retry_delay_seconds`  | Seconds to wait between retries            |
| `TimeoutSeconds`    | `int`       | `timeout_seconds`      | Maximum execution time before cancellation |
| `CreatedAt`         | `time.Time` | `created_at`           | Creation timestamp                         |

#### `TaskDependency`
Declares that a task must wait for another task to succeed first.

| Field             | Type        | JSON key             | Description                      |
|-------------------|-------------|----------------------|----------------------------------|
| `ID`              | `uuid.UUID` | `id`                 | Unique dependency identifier     |
| `TaskID`          | `uuid.UUID` | `task_id`            | The downstream (waiting) task    |
| `DependsOnTaskID` | `uuid.UUID` | `depends_on_task_id` | The upstream (prerequisite) task |

#### `WorkflowRun`
A single execution instance of a `Workflow`.

| Field        | Type         | JSON key      | Description                       |
|--------------|--------------|---------------|-----------------------------------|
| `ID`         | `uuid.UUID`  | `id`          | Unique run identifier             |
| `WorkflowID` | `uuid.UUID`  | `workflow_id` | The workflow being executed       |
| `Status`     | `Status`     | `status`      | Current lifecycle status          |
| `StartedAt`  | `time.Time`  | `started_at`  | When the run began                |
| `FinishedAt` | `*time.Time` | `finished_at` | When the run completed (nullable) |

#### `TaskRun`
A single execution attempt of a `Task` within a `WorkflowRun`.

| Field           | Type         | JSON key          | Description                           |
|-----------------|--------------|-------------------|---------------------------------------|
| `ID`            | `uuid.UUID`  | `id`              | Unique task run identifier            |
| `WorkflowRunID` | `uuid.UUID`  | `workflow_run_id` | Parent workflow run                   |
| `TaskID`        | `uuid.UUID`  | `task_id`         | The task being executed               |
| `Status`        | `Status`     | `status`          | Current lifecycle status              |
| `Attempt`       | `int`        | `attempt`         | Attempt number (1-based)              |
| `StartedAt`     | `time.Time`  | `started_at`      | When the attempt began                |
| `FinishedAt`    | `*time.Time` | `finished_at`     | When the attempt completed (nullable) |
| `Logs`          | `string`     | `logs`            | Captured stdout/stderr                |

#### `Worker`
A node that polls for and executes tasks.

| Field           | Type           | JSON key         | Description                     |
|-----------------|----------------|------------------|---------------------------------|
| `ID`            | `uuid.UUID`    | `id`             | Unique worker identifier        |
| `Hostname`      | `string`       | `hostname`       | Network hostname of the worker  |
| `LastHeartbeat` | `time.Time`    | `last_heartbeat` | Most recent heartbeat timestamp |
| `Status`        | `WorkerStatus` | `status`         | `active` or `inactive`          |

---

## Scheduler Interfaces (`domain/`)

The `domain/` package defines the port interfaces and lightweight scheduling entities
used by the service layer.

| File | Contents |
|------|----------|
| `domain/task.go` | `Task` entity, `TaskStatus` constants, `Priority` levels, `Validate()`, `CanRetry()`, `IsTerminal()` |
| `domain/worker.go` | `Worker` entity, `WorkerStatus` constants, `Validate()`, `HasCapacity()`, `IsAlive()` |
| `domain/interfaces.go` | `TaskRepository`, `WorkerRepository`, `Queue`, `Scheduler` interfaces |
| `domain/errors.go` | Sentinel errors: `ErrTaskNotFound`, `ErrWorkerNotFound`, `ErrQueueEmpty`, etc. |

---

## Test Structure

### `internal/domain/domain_test.go` (16 tests)

| Test name                           | What it covers                                                        |
|-------------------------------------|-----------------------------------------------------------------------|
| `TestStatusConstants`               | Status enum string values                                             |
| `TestWorkerStatusConstants`         | WorkerStatus enum string values                                       |
| `TestWorkflowInstantiation`         | Struct field assignment for Workflow                                  |
| `TestWorkflowJSONRoundtrip`         | JSON marshal/unmarshal round-trip for Workflow                        |
| `TestTaskInstantiation`             | Struct field assignment for Task                                      |
| `TestTaskJSONRoundtrip`             | JSON marshal/unmarshal round-trip for Task                            |
| `TestTaskDependencyInstantiation`   | Struct field assignment for TaskDependency                            |
| `TestTaskDependencyJSONRoundtrip`   | JSON marshal/unmarshal round-trip for TaskDependency                  |
| `TestWorkflowRunInstantiation`      | Struct field assignment; nil FinishedAt for running run               |
| `TestWorkflowRunOptionalFinishedAt` | `finished_at` omitted from JSON when nil                             |
| `TestWorkflowRunWithFinishedAt`     | `finished_at` preserved in JSON when set                             |
| `TestTaskRunInstantiation`          | Struct field assignment for TaskRun                                   |
| `TestTaskRunOptionalFinishedAt`     | `finished_at` omitted from JSON when nil                             |
| `TestWorkerInstantiation`           | Struct field assignment for Worker                                    |
| `TestWorkerJSONRoundtrip`           | JSON marshal/unmarshal round-trip for Worker                          |
| `TestStatusJSONField`               | Status enum serialises to correct JSON string value                   |

### `domain/domain_test.go` (18 tests)

| Test name                           | What it covers                                                        |
|-------------------------------------|-----------------------------------------------------------------------|
| `TestTask_Validate_*`               | Validation rules for Task (ID, Name, Priority, MaxRetries)            |
| `TestTask_CanRetry`                 | Retry eligibility when RetryCount < MaxRetries                        |
| `TestTask_IsTerminal`               | Terminal state detection (succeeded/failed)                           |
| `TestWorker_Validate_*`             | Validation rules for Worker (ID, Address, Concurrency)               |
| `TestWorker_HasCapacity_*`          | Capacity logic for idle/busy/offline workers                          |
| `TestWorker_IsAlive`                | Liveness check against heartbeat timeout                              |
| `TestSentinelErrors_NotNil`         | All sentinel errors are non-nil                                       |

---

## Database Migrations

Migration files live in `db/migrations/` and follow the `golang-migrate` naming convention.

### Prerequisites

Install [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate):

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Apply migrations (up)

```bash
migrate -path db/migrations -database "postgres://user:password@localhost:5432/mini_airflow?sslmode=disable" up
```

### Roll back migrations (down)

```bash
migrate -path db/migrations -database "postgres://user:password@localhost:5432/mini_airflow?sslmode=disable" down 1
```

---

## Database Schema

All tables use UUID primary keys generated by `uuid_generate_v4()` (requires the `uuid-ossp` PostgreSQL extension).

### `workflows`

| Column          | Type        | Constraints                  | Description                          |
|-----------------|-------------|------------------------------|--------------------------------------|
| `id`            | UUID        | PK, NOT NULL, DEFAULT uuid   | Unique workflow identifier           |
| `name`          | TEXT        | NOT NULL                     | Human-readable workflow name         |
| `description`   | TEXT        | NOT NULL, DEFAULT ''         | Optional description                 |
| `schedule_cron` | TEXT        | NOT NULL, DEFAULT ''         | Cron expression for scheduling       |
| `is_active`     | BOOLEAN     | NOT NULL, DEFAULT TRUE       | Whether the workflow is enabled      |
| `created_at`    | TIMESTAMPTZ | NOT NULL, DEFAULT NOW()      | Creation timestamp                   |

Indexes: `is_active`, `created_at`

### `tasks`

| Column                | Type        | Constraints                     | Description                                |
|-----------------------|-------------|---------------------------------|--------------------------------------------|
| `id`                  | UUID        | PK, NOT NULL, DEFAULT uuid      | Unique task identifier                     |
| `workflow_id`         | UUID        | NOT NULL, FK → workflows(id)    | Parent workflow                            |
| `name`                | TEXT        | NOT NULL                        | Task name                                  |
| `command`             | TEXT        | NOT NULL, DEFAULT ''            | Shell command or executable to run         |
| `retry_count`         | INT         | NOT NULL, DEFAULT 0             | Number of retry attempts on failure        |
| `retry_delay_seconds` | INT         | NOT NULL, DEFAULT 0             | Seconds to wait between retries            |
| `timeout_seconds`     | INT         | NOT NULL, DEFAULT 0             | Maximum execution time before cancellation |
| `created_at`          | TIMESTAMPTZ | NOT NULL, DEFAULT NOW()         | Creation timestamp                         |

Indexes: `workflow_id`, `created_at`

### `task_dependencies`

| Column              | Type | Constraints                          | Description                      |
|---------------------|------|--------------------------------------|----------------------------------|
| `id`                | UUID | PK, NOT NULL, DEFAULT uuid           | Unique dependency identifier     |
| `task_id`           | UUID | NOT NULL, FK → tasks(id)             | The downstream (waiting) task    |
| `depends_on_task_id`| UUID | NOT NULL, FK → tasks(id)             | The upstream (prerequisite) task |

Unique constraint: `(task_id, depends_on_task_id)`. Indexes: `task_id`, `depends_on_task_id`

### `workflow_runs`

| Column        | Type        | Constraints                      | Description                       |
|---------------|-------------|----------------------------------|-----------------------------------|
| `id`          | UUID        | PK, NOT NULL, DEFAULT uuid       | Unique run identifier             |
| `workflow_id` | UUID        | NOT NULL, FK → workflows(id)     | The workflow being executed       |
| `status`      | TEXT        | NOT NULL, DEFAULT 'pending'      | Current lifecycle status          |
| `started_at`  | TIMESTAMPTZ | NOT NULL, DEFAULT NOW()          | When the run began                |
| `finished_at` | TIMESTAMPTZ | NULL                             | When the run completed (nullable) |

Indexes: `workflow_id`, `status`, `started_at`

### `task_runs`

| Column            | Type        | Constraints                           | Description                           |
|-------------------|-------------|---------------------------------------|---------------------------------------|
| `id`              | UUID        | PK, NOT NULL, DEFAULT uuid            | Unique task run identifier            |
| `workflow_run_id` | UUID        | NOT NULL, FK → workflow_runs(id)      | Parent workflow run                   |
| `task_id`         | UUID        | NOT NULL, FK → tasks(id)              | The task being executed               |
| `status`          | TEXT        | NOT NULL, DEFAULT 'pending'           | Current lifecycle status              |
| `attempt`         | INT         | NOT NULL, DEFAULT 1                   | Attempt number (1-based)              |
| `started_at`      | TIMESTAMPTZ | NOT NULL, DEFAULT NOW()               | When the attempt began                |
| `finished_at`     | TIMESTAMPTZ | NULL                                  | When the attempt completed (nullable) |
| `logs`            | TEXT        | NOT NULL, DEFAULT ''                  | Captured stdout/stderr                |

Indexes: `workflow_run_id`, `task_id`, `status`, `started_at`

### `workers`

| Column           | Type        | Constraints                  | Description                     |
|------------------|-------------|------------------------------|---------------------------------|
| `id`             | UUID        | PK, NOT NULL, DEFAULT uuid   | Unique worker identifier        |
| `hostname`       | TEXT        | NOT NULL                     | Network hostname of the worker  |
| `last_heartbeat` | TIMESTAMPTZ | NOT NULL, DEFAULT NOW()      | Most recent heartbeat timestamp |
| `status`         | TEXT        | NOT NULL, DEFAULT 'active'   | `active` or `inactive`          |

Indexes: `status`, `last_heartbeat`

---

## Project Roadmap

- [x] **Phase 1** — Domain models (`internal/domain`) + Scheduler interfaces (`domain/`)
  - Core structs: `Workflow`, `Task`, `TaskDependency`, `WorkflowRun`, `TaskRun`, `Worker`
  - Typed `Status` and `WorkerStatus` enums; UUID-based IDs
  - Scheduler port interfaces: `TaskRepository`, `WorkerRepository`, `Queue`, `Scheduler`
  - 34 unit tests (16 + 18) — all passing
- [x] **Phase 2** — SQL migration files (`db/migrations/`)
  - `000001_init.up.sql` — creates all six tables with UUID PKs, FK constraints, and indexes
  - `000001_init.down.sql` — drops all tables in reverse dependency order
  - README updated with migration instructions and full DB schema reference
- [ ] **Phase 3** — Repository interfaces & in-memory implementations
  - Thread-safe `TaskRepository` and `WorkerRepository` backed by `sync.RWMutex` maps
  - Priority `Queue` using `container/heap` + `sync.Cond`
  - `SchedulerService` use-case wiring them together
- [ ] **Phase 4** — Scheduler service (cron-based workflow triggering)
- [ ] **Phase 5** — Worker service (task execution, heartbeat, retry logic)
- [ ] **Phase 6** — REST API (workflow/run management endpoints)
- [ ] **Phase 7** — Persistence layer (PostgreSQL — wire up repositories to real DB)
- [ ] **Phase 8** — Observability (metrics, structured logging, tracing)
