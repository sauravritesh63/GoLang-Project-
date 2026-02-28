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
   internal/repository/          ← Phase 3 ✅ (repository interfaces + GORM + mocks)
   ├── interfaces.go             (WorkflowRepository, TaskRepository, …)
   ├── postgres/                 (GORM-backed implementations)
   └── mock/                     (in-memory implementations for testing)
        │
        ▼
   internal/api/                ← Phase 4 ✅ (REST API + WebSocket hub)
   ├── router.go                (Gin engine wiring — dependency injection)
   ├── service/service.go       (business-logic layer — context-aware)
   ├── handler/handler.go       (HTTP handlers)
   └── websocket/hub.go         (real-time event broadcasting)
        │
        ▼
   scheduler/                   ← Phase 5 ✅ (Scheduler + in-memory Queue)
   ├── queue.go                 (MemQueue — thread-safe, unbounded FIFO)
   └── scheduler.go             (Scheduler — Submit, Cancel, Status)
        │
        ▼
   worker/                      ← Phase 6 ✅ (Worker service — execute, heartbeat, retry)
   └── worker.go                (Worker — dequeue, execute, retry, heartbeat loop)
        │
        ▼
   observability/               ← Phase 7 ✅ (structured logging + Prometheus metrics)
   ├── logging/logging.go       (zerolog-based structured logger with workflow/task context)
   └── metrics/metrics.go       (Prometheus counters, histograms, gauges)
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

### `internal/repository/mock/mock_test.go` (30 tests)

| Test name                                    | What it covers                                              |
|----------------------------------------------|-------------------------------------------------------------|
| `TestWorkflowRepo_CreateAndGetByID`          | Create + round-trip fetch                                   |
| `TestWorkflowRepo_GetByID_NotFound`          | ErrNotFound on unknown ID                                   |
| `TestWorkflowRepo_Update`                    | Field mutation persists                                     |
| `TestWorkflowRepo_Update_NotFound`           | ErrNotFound when updating missing record                    |
| `TestWorkflowRepo_Delete`                    | Record is removed                                           |
| `TestWorkflowRepo_Delete_NotFound`           | ErrNotFound on second delete                                |
| `TestWorkflowRepo_List`                      | All records returned                                        |
| `TestWorkflowRepo_ListActive`                | Only active workflows returned                              |
| `TestTaskRepo_CreateAndGetByID`              | Create + round-trip fetch                                   |
| `TestTaskRepo_GetByID_NotFound`              | ErrNotFound on unknown ID                                   |
| `TestTaskRepo_ListByWorkflowID`              | Filters by workflow_id correctly                            |
| `TestTaskRepo_Delete`                        | Record removed; second delete → ErrNotFound                 |
| `TestWorkflowRunRepo_CreateAndGetByID`       | Create + round-trip fetch                                   |
| `TestWorkflowRunRepo_UpdateStatus`           | Status + FinishedAt updated atomically                      |
| `TestWorkflowRunRepo_UpdateStatus_NotFound`  | ErrNotFound on unknown ID                                   |
| `TestWorkflowRunRepo_ListByWorkflowID`       | Filters by workflow_id correctly                            |
| `TestWorkflowRunRepo_ListByStatus`           | Filters by status correctly                                 |
| `TestTaskRunRepo_CreateAndGetByID`           | Create + round-trip fetch                                   |
| `TestTaskRunRepo_UpdateStatus`               | Status updated                                              |
| `TestTaskRunRepo_ListByWorkflowRunID`        | Filters by workflow_run_id correctly                        |
| `TestTaskRunRepo_ListByTaskID`               | Filters by task_id correctly                                |
| `TestTaskRunRepo_ListByStatus`               | Filters by status correctly                                 |
| `TestWorkerRepo_CreateAndGetByID`            | Create + round-trip fetch                                   |
| `TestWorkerRepo_GetByID_NotFound`            | ErrNotFound on unknown ID                                   |
| `TestWorkerRepo_Update`                      | Field mutation persists                                     |
| `TestWorkerRepo_Delete`                      | Record removed; second delete → ErrNotFound                 |
| `TestWorkerRepo_ListActive`                  | Only active workers returned                                |
| `TestWorkerRepo_UpdateHeartbeat`             | LastHeartbeat updated                                       |
| `TestWorkerRepo_UpdateHeartbeat_NotFound`    | ErrNotFound on unknown ID                                   |
| _(compile-time interface checks)_            | All mock types satisfy repository interfaces                |

### `internal/repository/postgres/postgres_test.go`

Compile-time `var _ repository.XxxRepository = (*postgres.XxxRepo)(nil)` checks
verify that each GORM implementation satisfies the corresponding interface.

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

### Migration verification status

| Item | Status | Notes |
|------|--------|-------|
| `db/migrations/` directory | ✅ Present | |
| `000001_init.up.sql` | ✅ Present | Creates all six tables, indexes, and `uuid-ossp` extension |
| `000001_init.down.sql` | ✅ Present | Drops all tables in reverse dependency order |
| `workflows` table | ✅ Matches domain | Columns: `id`, `name`, `description`, `schedule_cron`, `is_active`, `created_at` |
| `tasks` table | ✅ Matches domain | Columns: `id`, `workflow_id`, `name`, `command`, `retry_count`, `retry_delay_seconds`, `timeout_seconds`, `created_at` |
| `task_dependencies` table | ✅ Matches domain | Columns: `id`, `task_id`, `depends_on_task_id`; unique on `(task_id, depends_on_task_id)` |
| `workflow_runs` table | ✅ Matches domain | Columns: `id`, `workflow_id`, `status`, `started_at`, `finished_at` |
| `task_runs` table | ✅ Matches domain | Columns: `id`, `workflow_run_id`, `task_id`, `status`, `attempt`, `started_at`, `finished_at`, `logs` |
| `workers` table | ✅ Matches domain | Columns: `id`, `hostname`, `last_heartbeat`, `status` |
| Indexes | ✅ All present | See [Database Schema](#database-schema) for full index list |

Nothing is missing — the migration files and schema are complete and consistent with the domain specification.

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

## Repository Layer (`internal/repository`)

Phase 3 adds a repository layer that decouples the service/use-case code from the
database technology. The **repository pattern** used here consists of three parts:

| Package | Purpose |
|---------|---------|
| `internal/repository` | Interface definitions only — the contract every concrete implementation must honour |
| `internal/repository/postgres` | GORM-backed implementations for PostgreSQL |
| `internal/repository/mock` | Thread-safe in-memory implementations for unit testing |

### Repository Interfaces

All interfaces live in `internal/repository/interfaces.go`. Every method accepts a
`context.Context` as its first argument so callers can propagate deadlines and
cancellation signals to the database.

#### `WorkflowRepository`

```go
type WorkflowRepository interface {
    Create(ctx context.Context, wf *domain.Workflow) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Workflow, error)
    Update(ctx context.Context, wf *domain.Workflow) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context) ([]*domain.Workflow, error)
    ListActive(ctx context.Context) ([]*domain.Workflow, error)
}
```

#### `TaskRepository`

```go
type TaskRepository interface {
    Create(ctx context.Context, t *domain.Task) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
    Update(ctx context.Context, t *domain.Task) error
    Delete(ctx context.Context, id uuid.UUID) error
    ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.Task, error)
}
```

#### `WorkflowRunRepository`

```go
type WorkflowRunRepository interface {
    Create(ctx context.Context, wr *domain.WorkflowRun) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowRun, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, finishedAt *time.Time) error
    ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.WorkflowRun, error)
    ListByStatus(ctx context.Context, status domain.Status) ([]*domain.WorkflowRun, error)
}
```

#### `TaskRunRepository`

```go
type TaskRunRepository interface {
    Create(ctx context.Context, tr *domain.TaskRun) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.TaskRun, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, finishedAt *time.Time) error
    ListByWorkflowRunID(ctx context.Context, workflowRunID uuid.UUID) ([]*domain.TaskRun, error)
    ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskRun, error)
    ListByStatus(ctx context.Context, status domain.Status) ([]*domain.TaskRun, error)
}
```

#### `WorkerRepository`

```go
type WorkerRepository interface {
    Create(ctx context.Context, w *domain.Worker) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Worker, error)
    Update(ctx context.Context, w *domain.Worker) error
    Delete(ctx context.Context, id uuid.UUID) error
    ListActive(ctx context.Context) ([]*domain.Worker, error)
    UpdateHeartbeat(ctx context.Context, id uuid.UUID, at time.Time) error
}
```

### Sentinel error

`repository.ErrNotFound` is returned by any method when the requested record does
not exist. Callers can test for it with `errors.Is`.

### GORM (PostgreSQL) implementation

`internal/repository/postgres` provides one concrete struct per interface (e.g.
`WorkflowRepo`, `TaskRepo`). Each struct accepts a `*gorm.DB` via its constructor:

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

workflowRepo := postgresrepo.NewWorkflowRepo(db)
taskRepo     := postgresrepo.NewTaskRepo(db)
// …
```

Internally each struct maintains private GORM model types (with `gorm:` struct
tags and a `TableName()` method) and converts them to/from the domain types.
This keeps the domain package free of any ORM or database concern.

### Mock (in-memory) implementation

`internal/repository/mock` provides thread-safe in-memory implementations backed
by `sync.RWMutex`-protected maps. They are designed for unit tests and satisfy
the same interfaces:

```go
workflowRepo := mock.NewWorkflowRepo()
taskRepo     := mock.NewTaskRepo()
// …
```

### Design principles

- **Dependency injection** — every concrete repo receives its dependencies
  (`*gorm.DB`) via a `NewXxx(db)` constructor; there are no package-level global
  variables.
- **Context-aware** — every method signature starts with `context.Context`,
  which is forwarded to `db.WithContext(ctx)` so timeouts and cancellations are
  always respected.
- **Interface segregation** — one small interface per aggregate root; callers
  only depend on what they actually use.
- **Testability** — business logic can be tested without a live database by
  injecting a `mock.*Repo` instead of a `postgres.*Repo`.

---

## REST API (`internal/api`)

Phase 4 adds a Gin-based HTTP API with the following endpoints.

### Service Architecture

```
HTTP Client
    │
    ▼
internal/api/router.go      ← Gin engine; dependency injection entry point
    │
    ▼
internal/api/handler/       ← HTTP layer: parse request, call service, write JSON
    │
    ▼
internal/api/service/       ← Business logic: orchestrate repositories
    │
    ▼
internal/repository/        ← Repository interfaces (Phase 3)
```

The WebSocket hub (`internal/api/websocket/Hub`) is injected into the handler
layer and receives `Broadcast` calls whenever a workflow is triggered.

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/workflows` | Create a new workflow |
| `GET`  | `/workflows` | List workflows (paginated) |
| `POST` | `/workflows/{id}/trigger` | Trigger a new run of a workflow |
| `GET`  | `/workflow-runs` | List workflow runs (optional `?status=` filter) |
| `GET`  | `/task-runs` | List task runs (optional `?status=` filter) |
| `GET`  | `/workers` | List active workers |
| `GET`  | `/ws/updates` | WebSocket — real-time event stream |

#### Pagination

`GET /workflows` supports `?offset=<int>&limit=<int>` query parameters.

| Parameter | Default | Description |
|-----------|---------|-------------|
| `offset`  | `0`     | Number of records to skip |
| `limit`   | `20`    | Maximum number of records to return |

#### Status filter

`GET /workflow-runs` and `GET /task-runs` accept an optional `?status=` query
parameter. Valid values: `pending`, `running`, `success`, `failed`.

### Example curl Usage

```bash
# Create a workflow
curl -s -X POST http://localhost:8080/workflows \
  -H 'Content-Type: application/json' \
  -d '{"name":"daily-etl","description":"Daily ETL pipeline","schedule_cron":"0 2 * * *","is_active":true}' | jq

# List workflows (first page)
curl -s 'http://localhost:8080/workflows?offset=0&limit=10' | jq

# Trigger a workflow run (replace <id> with the UUID from the create response)
curl -s -X POST http://localhost:8080/workflows/<id>/trigger | jq

# List workflow runs filtered by status
curl -s 'http://localhost:8080/workflow-runs?status=pending' | jq

# List task runs
curl -s 'http://localhost:8080/task-runs' | jq

# List active workers
curl -s 'http://localhost:8080/workers' | jq
```

### WebSocket Usage

Connect to `ws://localhost:8080/ws/updates` to receive real-time JSON events.

**Event envelope:**

```json
{
  "type": "workflow_status",
  "payload": { ... }
}
```

| `type` value | Emitted when |
|---|---|
| `workflow_status` | A workflow run is created / its status changes |
| `task_status` | A task run changes state |
| `worker_heartbeat` | A worker sends a heartbeat |

**Example — connect with `websocat`:**

```bash
websocat ws://localhost:8080/ws/updates
```

**Example — connect with JavaScript (browser):**

```js
const ws = new WebSocket('ws://localhost:8080/ws/updates');
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log(msg.type, msg.payload);
};
```

### Starting the API server

The router is created via `api.NewRouter` with injected repository
implementations. Wire it up in your `main.go`:

```go
package main

import (
    "github.com/sauravritesh63/GoLang-Project-/internal/api"
    "github.com/sauravritesh63/GoLang-Project-/internal/repository/mock"
)

func main() {
    // Replace mock repos with postgres implementations when a DB is available.
    r := api.NewRouter(
        mock.NewWorkflowRepo(),
        mock.NewWorkflowRunRepo(),
        mock.NewTaskRunRepo(),
        mock.NewWorkerRepo(),
    )
    r.Run(":8080")
}
```

---

## Scheduler (`scheduler/`)

Phase 5 adds the core scheduling primitives used by the Worker service.

### MemQueue

`scheduler.MemQueue` is a thread-safe, unbounded, FIFO in-memory queue that satisfies the `domain.Queue` interface.

```go
q := scheduler.NewMemQueue()

// Enqueue a task.
_ = q.Enqueue(ctx, task)

// Dequeue blocks until a task is available or ctx is cancelled.
t, err := q.Dequeue(ctx)

// Len returns the current queue depth.
n, _ := q.Len(ctx)
```

### Scheduler

`scheduler.Scheduler` satisfies the `domain.Scheduler` interface and orchestrates task submission, cancellation, and status queries.

```go
sched := scheduler.New(taskRepo, workerRepo, queue)

// Submit validates the task, marks it Queued, persists it, and enqueues it.
_ = sched.Submit(ctx, task)

// Cancel marks a non-terminal task as Failed (no-op for terminal tasks).
_ = sched.Cancel(ctx, task.ID)

// Status returns the current TaskStatus.
status, _ := sched.Status(ctx, task.ID)
```

---

## Worker Service (`worker/`)

Phase 6 adds the Worker service that pulls tasks from the queue and executes them.

### Worker

`worker.Worker` registers itself with the `WorkerRepository`, processes tasks one at a time, retries failed tasks up to `task.MaxRetries` times with exponential backoff, and sends periodic heartbeats.

```go
// handler is your business logic for executing a task payload.
// Use worker.MockShellHandler during development / testing.
handler := func(ctx context.Context, task *domain.Task) error {
    // ... process task.Payload ...
    return nil
}

w := worker.New(
    "worker-1",        // unique worker ID / address
    queue,             // domain.Queue (e.g. scheduler.NewMemQueue())
    taskRepo,          // domain.TaskRepository
    workerRepo,        // domain.WorkerRepository
    handler,
    worker.WithHeartbeatInterval(15*time.Second), // optional; default 15 s
    worker.WithBackoff(worker.DefaultBackoff),     // optional; default exponential
)

// Run blocks until ctx is cancelled.
if err := w.Run(ctx); err != nil {
    log.Fatal(err)
}
```

#### Configuration options

| Option | Default | Description |
|--------|---------|-------------|
| `WithHeartbeatInterval(d)` | 15 s | How often the worker refreshes its `LastHeartAt` timestamp in the `WorkerRepository`. |
| `WithBackoff(fn)` | `DefaultBackoff` | Function that returns the delay before each retry attempt. `DefaultBackoff` gives 1 s, 2 s, 4 s … capped at 30 s. Pass `func(int) time.Duration { return 0 }` in tests for instant retries. |

#### MockShellHandler

`worker.MockShellHandler` is a built-in `Handler` that simulates shell-command execution using the task's `Payload` field. It always succeeds and is suitable for development and unit tests before a real executor is wired in.

```go
w := worker.New("worker-1", queue, taskRepo, workerRepo, worker.MockShellHandler)
```

#### Task lifecycle managed by the worker

| Transition | Condition |
|------------|-----------|
| `queued` → `running` | Task dequeued |
| `running` → `succeeded` | Handler returned nil |
| `running` → `retrying` | Handler returned error **and** `task.CanRetry()` is true |
| `retrying` → `running` | Backoff delay elapsed; task re-enqueued and dequeued again |
| `running` → `failed` | Handler returned error **and** no retries remaining |

#### Deployment

Run one or more workers alongside the API server. Each worker is stateless — scale horizontally by starting additional processes with unique IDs:

```bash
# Terminal 1 — API server
go run ./cmd/api

# Terminal 2 — worker-1
WORKER_ID=worker-1 go run ./cmd/worker

# Terminal 3 — worker-2 (horizontal scale-out)
WORKER_ID=worker-2 go run ./cmd/worker
```

Workers register themselves in the `WorkerRepository` on startup and send heartbeats at the configured interval. Remove a worker at any time by cancelling its context (e.g., `SIGTERM`); in-flight tasks will reach a terminal state before the goroutine exits.

---

## Observability (`observability/`)

Phase 7 adds centralized structured logging, Prometheus metrics, and HTTP health/metrics endpoints.

### Structured Logging (`observability/logging`)

The `logging` package wraps [zerolog](https://github.com/rs/zerolog) and provides:

- A package-level `Logger` (JSON to stdout) ready to use with zero configuration.
- `New(w io.Writer)` — create a logger writing to any `io.Writer`.
- `WithContext` / `FromContext` — embed a logger in a `context.Context` and retrieve it anywhere in a call chain.
- `WithWorkflow`, `WithTask`, `WithWorker` — attach contextual fields so every log line carries `workflow_id`, `task_id`, or `worker_id`.

```go
import "github.com/sauravritesh63/GoLang-Project-/observability/logging"

// Use the default logger (JSON, stdout).
logging.Logger.Info().Str("event", "server_start").Int("port", 8080).Msg("API server listening")

// Attach a workflow-scoped logger to a context.
ctx = logging.WithContext(ctx, logging.WithWorkflow(logging.Logger, wf.ID.String(), wf.Name))

// Retrieve it later in any function that receives ctx.
log := logging.FromContext(ctx)
log.Info().Str("status", "triggered").Msg("workflow run started")
```

All log lines are valid JSON and include a `time` field (RFC3339). Pipe output to `jq` or any log-aggregation platform (Loki, Datadog, CloudWatch, etc.).

### Prometheus Metrics (`observability/metrics`)

Call `metrics.New()` once during startup to register all counters and histograms with the default Prometheus registry:

```go
import "github.com/sauravritesh63/GoLang-Project-/observability/metrics"

col := metrics.New()

// Increment after each workflow run is triggered.
col.WorkflowsTotal.WithLabelValues("pending").Inc()

// Record task execution duration.
col.TaskDuration.WithLabelValues("succeeded").Observe(duration.Seconds())

// Count heartbeats per worker.
col.WorkerHeartbeats.WithLabelValues(workerID).Inc()

// Count retries per worker.
col.TaskRetries.WithLabelValues(workerID).Inc()
```

#### Metrics reference

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `scheduler_workflows_total` | Counter | `status` | Total workflow runs triggered |
| `scheduler_tasks_total` | Counter | `status` | Total task runs processed |
| `scheduler_task_duration_seconds` | Histogram | `status` | Task execution duration |
| `scheduler_workflow_failures_total` | Counter | — | Total workflow run failures |
| `scheduler_workflow_successes_total` | Counter | — | Total workflow run successes |
| `scheduler_worker_heartbeats_total` | Counter | `worker_id` | Total worker heartbeat ticks |
| `scheduler_task_retries_total` | Counter | `worker_id` | Total task retry attempts |

### HTTP Endpoints

| Service   | Endpoint | Method | Description |
|-----------|----------|--------|-------------|
| api       | `/metrics` | GET | Prometheus scrape endpoint — exposes all registered metrics in text format |
| api       | `/healthz` | GET | Health check — returns `{"status":"ok","service":"task-scheduler-api"}` |
| scheduler | `/metrics` | GET | Prometheus scrape endpoint (port `METRICS_PORT`, default `9090`) |
| scheduler | `/healthz` | GET | Health check — returns `{"status":"ok","service":"task-scheduler-scheduler"}` |
| worker    | `/metrics` | GET | Prometheus scrape endpoint (port `METRICS_PORT`, default `9091`) |
| worker    | `/healthz` | GET | Health check — returns `{"status":"ok","service":"task-scheduler-worker"}` |

**Example — check health:**
```bash
curl http://localhost:8080/healthz
# {"service":"task-scheduler-api","status":"ok"}
```

**Example — scrape metrics:**
```bash
curl http://localhost:8080/metrics
# HELP scheduler_workflows_total Total number of workflow runs triggered.
# TYPE scheduler_workflows_total counter
# scheduler_workflows_total{status="pending"} 3
# ...
```

### Prometheus Setup (docker-compose snippet)

Add the following to your `docker-compose.yml` to scrape all three services:

```yaml
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"
```

`prometheus.yml`:
```yaml
scrape_configs:
  - job_name: task-scheduler-api
    static_configs:
      - targets: ["api:8080"]
  - job_name: task-scheduler-scheduler
    static_configs:
      - targets: ["scheduler:9090"]
  - job_name: task-scheduler-worker
    static_configs:
      - targets: ["worker:9091"]
```

> **Kubernetes:** The `scheduler` and `worker` Deployments already include `prometheus.io/scrape: "true"` pod annotations, so a standard Prometheus operator or annotation-based scrape config will pick them up automatically.

---

## Phase 8 — Containerization, CI/CD & Kubernetes

### Local Development with Docker Compose

```bash
# Build and start all services (Postgres, Redis, API, Scheduler, Worker)
docker compose up --build

# API is available at http://localhost:8080
curl http://localhost:8080/healthz
# {"service":"task-scheduler-api","status":"ok"}

# Tear down
docker compose down -v
```

**Services started by `docker compose up`:**

| Service   | Port | Description |
|-----------|------|-------------|
| postgres  | 5432 | PostgreSQL 16 (schema auto-applied via init scripts) |
| redis     | 6379 | Redis 7 (for future queue integration) |
| api       | 8080 | REST API + Prometheus metrics |
| scheduler | —    | Scheduler service |
| worker    | —    | Task worker (WORKER_ID=worker-1) |

### Dockerfiles

Each service has its own multi-stage Dockerfile under `docker/<service>/Dockerfile`:

| Dockerfile | Build target | Base image |
|------------|-------------|------------|
| `docker/api/Dockerfile` | `cmd/api` | `gcr.io/distroless/static-debian12:nonroot` |
| `docker/scheduler/Dockerfile` | `cmd/scheduler` | `gcr.io/distroless/static-debian12:nonroot` |
| `docker/worker/Dockerfile` | `cmd/worker` | `gcr.io/distroless/static-debian12:nonroot` |

Build images manually:
```bash
docker build -f docker/api/Dockerfile -t task-scheduler-api:dev .
docker build -f docker/scheduler/Dockerfile -t task-scheduler-scheduler:dev .
docker build -f docker/worker/Dockerfile -t task-scheduler-worker:dev .
```

**Environment variables:**

| Variable | Service | Default | Description |
|----------|---------|---------|-------------|
| `PORT` | api | `8080` | HTTP listen port |
| `DATABASE_URL` | api | `""` | PostgreSQL DSN (in-memory fallback if unset) |
| `GIN_MODE` | api | `release` | Gin mode (`debug`/`release`) |
| `WORKER_ID` | worker | `worker-1` | Unique worker identifier |
| `METRICS_PORT` | scheduler | `9090` | Port for `/metrics` and `/healthz` endpoints |
| `METRICS_PORT` | worker | `9091` | Port for `/metrics` and `/healthz` endpoints |
| `LOG_LEVEL` | all | `info` | Log verbosity |

### CI/CD Pipelines (GitHub Actions)

#### `.github/workflows/ci.yaml` — Continuous Integration

Triggered on every push and pull request to `main`/`master`:

1. **Lint** — `golangci-lint` enforces code quality
2. **Test** — `go test -race ./...` with coverage upload to Codecov
3. **Build** — compiles all three binaries and validates Docker images

#### `.github/workflows/release.yaml` — Release

Triggered on `v*.*.*` tag push:

1. Runs full test suite
2. Cross-compiles binaries for `linux/amd64` and `linux/arm64`
3. Builds and pushes multi-arch Docker images to GHCR
4. Creates a GitHub Release with auto-generated release notes and binary attachments

**To cut a release:**
```bash
git tag v1.0.0
git push origin v1.0.0
```

Images are published to:
- `ghcr.io/<owner>/task-scheduler-api:<tag>`
- `ghcr.io/<owner>/task-scheduler-scheduler:<tag>`
- `ghcr.io/<owner>/task-scheduler-worker:<tag>`

### Kubernetes — Production Deployment

Apply manifests in order:
```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml   # update DATABASE_URL secret first!
kubectl apply -f k8s/api-deployment.yaml
kubectl apply -f k8s/scheduler-deployment.yaml
kubectl apply -f k8s/worker-deployment.yaml
```

**Manifest summary:**

| File | Resource | Replicas | Notes |
|------|----------|----------|-------|
| `k8s/namespace.yaml` | Namespace | — | `task-scheduler` namespace |
| `k8s/configmap.yaml` | ConfigMap + Secret | — | Non-secret config + DATABASE_URL |
| `k8s/api-deployment.yaml` | Deployment + Service | 2 | Liveness & readiness probes on `/healthz` |
| `k8s/scheduler-deployment.yaml` | Deployment | 1 | Single leader; scale with leader-election if needed |
| `k8s/worker-deployment.yaml` | Deployment | 3 | Pod name injected as `WORKER_ID`; scale freely |

**Scaling workers:**
```bash
kubectl -n task-scheduler scale deployment worker --replicas=10
```

**Health and readiness probes (API):**
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: http
  initialDelaySeconds: 10
  periodSeconds: 15

readinessProbe:
  httpGet:
    path: /healthz
    port: http
  initialDelaySeconds: 5
  periodSeconds: 10
```

> **Security note:** Replace the placeholder `DATABASE_URL` in `k8s/configmap.yaml` with a
> reference to an external secrets manager (e.g. AWS Secrets Manager, HashiCorp Vault, or
> the Kubernetes External Secrets Operator) before deploying to production.

---



- [x] **Phase 1** — Domain models (`internal/domain`) + Scheduler interfaces (`domain/`)
  - Core structs: `Workflow`, `Task`, `TaskDependency`, `WorkflowRun`, `TaskRun`, `Worker`
  - Typed `Status` and `WorkerStatus` enums; UUID-based IDs
  - Scheduler port interfaces: `TaskRepository`, `WorkerRepository`, `Queue`, `Scheduler`
  - 34 unit tests (16 + 18) — all passing
- [x] **Phase 2** — SQL migration files (`db/migrations/`)
  - `000001_init.up.sql` — creates all six tables with UUID PKs, FK constraints, and indexes
  - `000001_init.down.sql` — drops all tables in reverse dependency order
  - README updated with migration instructions and full DB schema reference
- [x] **Phase 3** — Repository interfaces & GORM implementations
  - `internal/repository/interfaces.go` — five typed interfaces (WorkflowRepository, TaskRepository, WorkflowRunRepository, TaskRunRepository, WorkerRepository)
  - `internal/repository/postgres/` — GORM-backed implementations (dependency injection, context-aware, no global state)
  - `internal/repository/mock/` — thread-safe in-memory implementations for unit testing
  - 29 unit tests in `mock/mock_test.go` — all passing; compile-time interface checks in `postgres/postgres_test.go`
- [x] **Phase 4** — REST API + WebSocket hub (`internal/api/`)
  - `internal/api/router.go` — Gin engine wired with dependency injection
  - `internal/api/service/` — business-logic layer (context-aware, no direct DB access)
  - `internal/api/handler/` — HTTP handlers for all six REST endpoints
  - `internal/api/websocket/` — Gorilla WebSocket hub for real-time event broadcasting
  - 10 unit tests in `handler/handler_test.go` — all passing
  - README updated with endpoint reference, WebSocket usage, and architecture notes
- [x] **Phase 5** — Scheduler service (`scheduler/`)
  - `scheduler/queue.go` — thread-safe, unbounded in-memory `MemQueue` implementing `domain.Queue` (FIFO, blocking Dequeue with context cancellation)
  - `scheduler/scheduler.go` — `Scheduler` struct implementing `domain.Scheduler` (Submit, Cancel, Status)
  - 14 unit tests in `scheduler/scheduler_test.go` — all passing; compile-time interface checks included
- [x] **Phase 6** — Worker service (`worker/`) — task execution, heartbeat, retry logic
  - `worker/worker.go` — `Worker` struct: registers with `WorkerRepository`, dequeues tasks, executes via pluggable `Handler`, transitions task status, retries on failure with exponential backoff, sends periodic heartbeats
  - `worker.MockShellHandler` — built-in handler that simulates shell-command execution for development and testing
  - `worker.DefaultBackoff` — exponential backoff (1 s, 2 s, 4 s … capped at 30 s); overridable via `WithBackoff`
  - `worker.WithBackoff` functional option for injecting a custom or zero-delay backoff (useful in tests)
  - 8 unit tests in `worker/worker_test.go` — all passing (register, success, retry, no-retry, clean shutdown, heartbeat, mock-handler, backoff timing)
- [x] **Phase 7** — Observability (structured logging, Prometheus metrics, health checks)
  - `observability/logging/logging.go` — zerolog-based structured logger with context propagation and per-workflow/task/worker field helpers
  - `observability/metrics/metrics.go` — Prometheus `Collector` with counters/histograms for workflows, tasks, workers, and retries
  - `GET /metrics` — Prometheus scrape endpoint (registered on the Gin router via `promhttp.Handler()`)
  - `GET /healthz` — Health check endpoint returning `{"status":"ok","service":"task-scheduler-api"}`
  - 1 unit test in `handler/handler_test.go` — `TestHealthz` — all passing
  - README updated with observability endpoints, metrics reference, and setup instructions
- [x] **Phase 8** — Containerization, CI/CD, and Kubernetes (`docker/`, `docker-compose.yaml`, `.github/workflows/`, `k8s/`)
  - Multi-stage Dockerfiles for each service (`docker/api/Dockerfile`, `docker/scheduler/Dockerfile`, `docker/worker/Dockerfile`)
  - `docker-compose.yaml` for full local orchestration (Postgres, Redis, API, Scheduler, Worker) with health checks on all services
  - GitHub Actions CI pipeline (`.github/workflows/ci.yaml`): lint → test → build → Docker image validation
  - GitHub Actions Release pipeline (`.github/workflows/release.yaml`): multi-arch binaries + GHCR images pushed on tag
  - Kubernetes manifests (`k8s/`) for production: Namespace, ConfigMap/Secret, API Deployment (2 replicas, liveness/readiness probes), Scheduler Deployment (1 replica, liveness/readiness on `:9090/healthz`), Worker Deployment (3 replicas, liveness/readiness on `:9091/healthz`, pod name as WORKER_ID)
  - Service entry points added: `cmd/api/main.go`, `cmd/scheduler/main.go`, `cmd/worker/main.go`
  - All three services expose `/metrics` (Prometheus) and `/healthz` (health check) endpoints
  - README updated with deployment, scaling, and CI/CD instructions (see sections below)

---

## Production Checklist

Use this checklist before every production launch or major deployment.

### Infrastructure

- [ ] PostgreSQL deployed with replication and automatic failover (e.g. RDS Multi-AZ, Cloud SQL HA)
- [ ] Redis deployed with persistence enabled (AOF or RDB) and a replica for failover
- [ ] TLS termination configured at the load-balancer or ingress level (HTTPS everywhere)
- [ ] Database connection pool sized appropriately (`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`)
- [ ] Database credentials stored in a secrets manager (AWS Secrets Manager, HashiCorp Vault, or Kubernetes External Secrets), **not** hardcoded in ConfigMap

### Application

- [ ] All three services (`api`, `scheduler`, `worker`) start without errors (`go build ./...` passes)
- [ ] All unit tests pass with the race detector (`go test -race ./...`)
- [ ] `GET /healthz` returns HTTP 200 for the API service
- [ ] `GET /metrics` returns valid Prometheus text exposition
- [ ] `GIN_MODE=release` is set for the API container in production
- [ ] `LOG_LEVEL` is set to `info` or `warn` (avoid `debug` in production)
- [ ] Worker `WORKER_ID` values are unique per replica (use Kubernetes `metadata.name` downward API — already configured in `k8s/worker-deployment.yaml`)
- [ ] Graceful shutdown is verified: SIGTERM drains in-flight tasks before exit (`terminationGracePeriodSeconds` ≥ longest expected task duration)

### Kubernetes

- [ ] Manifests applied in order: `namespace.yaml` → `configmap.yaml` → deployments
- [ ] `DATABASE_URL` Secret value replaced with real credentials before `kubectl apply`
- [ ] API liveness and readiness probes validated (both probe `/healthz`)
- [ ] HPA (Horizontal Pod Autoscaler) configured for the Worker deployment if traffic is variable
- [ ] Resource `requests` and `limits` tuned based on observed P99 CPU/memory from load tests
- [ ] Pod Disruption Budgets (PDB) set to ensure at least one API replica is always available during rolling updates
- [ ] Network Policies applied to restrict inter-service traffic to required paths only

### Observability

- [ ] Prometheus scrapes `api:8080/metrics` (add `scheduler` and `worker` metric endpoints if needed)
- [ ] Grafana dashboards created for: workflow run rate, task success/failure ratio, task execution duration P50/P99, worker heartbeat lag
- [ ] Alerting rules configured for: high task failure rate, worker heartbeat silence > 60 s, API error rate spike
- [ ] Structured JSON logs are being collected by the log aggregator (Loki, CloudWatch, Datadog, etc.)
- [ ] Log retention policy set (e.g. 30 days hot, 1 year cold)

### Security

- [ ] Container images built from `gcr.io/distroless/static-debian12:nonroot` (no shell, runs as non-root) — already configured
- [ ] Image vulnerability scan passes (e.g. Trivy, Snyk) with no critical/high findings
- [ ] API does not expose sensitive data in error responses (`GIN_MODE=release` suppresses stack traces)
- [ ] Database migrations tested on a staging database before running on production
- [ ] `000001_init.down.sql` is reviewed and available for emergency rollback

### CI/CD

- [ ] CI pipeline (`ci.yaml`) green on the release commit: lint → test → build → Docker image smoke test
- [ ] Release pipeline (`release.yaml`) triggered via a semver tag (e.g. `git tag v1.0.0 && git push origin v1.0.0`)
- [ ] Multi-arch Docker images (`linux/amd64`, `linux/arm64`) published to GHCR
- [ ] GitHub Release created with auto-generated release notes and binary attachments

---

## Final Recommendations

### Short-term (before v1.0 launch)

1. **Replace in-memory queue with Redis Streams or a persistent queue** — The current `scheduler/queue.go` is an in-memory FIFO. If the scheduler pod restarts, queued tasks are lost. Integrate Redis Streams (or a PostgreSQL-backed queue like `pgqueuer`) so tasks survive restarts.

2. **Add scheduler leader-election** — The `k8s/scheduler-deployment.yaml` runs a single replica. Until leader-election is implemented (e.g. via a Kubernetes Lease object), scaling the scheduler to >1 replica will cause duplicate task submissions. Keep `replicas: 1` until this is addressed.

3. ~~**Instrument the scheduler and worker with Prometheus metrics**~~ ✅ **Done** — `cmd/scheduler/main.go` and `cmd/worker/main.go` now call `metrics.New()` at startup and expose `/metrics` and `/healthz` endpoints on dedicated ports (`METRICS_PORT`, defaulting to `9090` and `9091` respectively). K8s manifests updated with liveness/readiness probes and `prometheus.io/scrape` annotations. `docker-compose.yaml` updated with health checks on the new ports.

4. **Add integration / smoke tests** — The test suite covers domain logic and repository mocks well (80+ unit tests), but there are no end-to-end tests that spin up the full stack. Add at least one smoke test using `docker compose up` that verifies: workflow created → task queued → worker picks it up → status transitions to `success`.

5. **Cron-based workflow triggering** — The scheduler compiles and the domain supports `ScheduleCron`, but the current `cmd/scheduler/main.go` does not parse cron expressions and trigger `WorkflowRun` creation on schedule. Add a cron library (e.g. `github.com/robfig/cron/v3`) and implement the trigger loop.

### Medium-term (post-launch hardening)

6. **DAG dependency enforcement** — `TaskDependency` is persisted in the DB but the worker does not currently check that upstream tasks are complete before executing a downstream task. Implement a DAG resolver in the scheduler or worker service.

7. **Distributed tracing** — Add OpenTelemetry instrumentation (`go.opentelemetry.io/otel`) with a Jaeger or Tempo backend for end-to-end trace visibility across API → Scheduler → Worker.

8. **Rate-limiting and authentication on the REST API** — The API is currently open. Add JWT or API-key middleware (e.g. `github.com/gin-contrib/jwt`) and rate-limiting (e.g. `golang.org/x/time/rate`) before exposing to the internet.

9. **WebSocket authentication** — The WebSocket hub at `/ws/tasks` is unauthenticated. Add a token-based handshake before upgrading connections.

10. **Horizontal worker auto-scaling** — Configure a Kubernetes HPA on the Worker deployment that scales on a custom metric (e.g. queue depth exported via a Prometheus adapter), so the worker pool grows automatically under load.


---

## Final Review Summary

This section captures the overall project readiness at the time of the final review pass.

### Phase Completion

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Domain models (`internal/domain`) + Scheduler interfaces (`domain/`) | ✅ Complete |
| 2 | SQL migrations (`db/migrations/`) | ✅ Complete |
| 3 | Repository interfaces, GORM implementations, mock implementations | ✅ Complete |
| 4 | REST API (Gin), WebSocket hub, service layer | ✅ Complete |
| 5 | Scheduler service (`scheduler/`) — queue, submit, cancel, status | ✅ Complete |
| 6 | Worker service (`worker/`) — dequeue, execute, retry, heartbeat | ✅ Complete |
| 7 | Observability — structured logging, Prometheus metrics, `/metrics` + `/healthz` on all three services | ✅ Complete |
| 8 | Containerization, CI/CD, Kubernetes | ✅ Complete |

### Test Coverage

| Package | Tests | Status |
|---------|-------|--------|
| `internal/domain` | 16 unit tests | ✅ Pass |
| `domain` | 18 unit tests | ✅ Pass |
| `internal/repository/mock` | 30 unit tests | ✅ Pass |
| `internal/repository/postgres` | Compile-time interface checks | ✅ Pass |
| `internal/api/handler` | 10 unit tests (incl. `/healthz`) | ✅ Pass |
| `internal/api/service` | Unit tests | ✅ Pass |
| `internal/api/websocket` | Unit tests | ✅ Pass |
| `scheduler` | 14 unit tests | ✅ Pass |
| `worker` | 8 unit tests | ✅ Pass |

Run the full test suite with the race detector:
```bash
go test -race ./...
```

### Metrics & Health Endpoints

| Service | Port | `/healthz` | `/metrics` |
|---------|------|-----------|-----------|
| api | 8080 | ✅ | ✅ |
| scheduler | 9090 (configurable via `METRICS_PORT`) | ✅ | ✅ |
| worker | 9091 (configurable via `METRICS_PORT`) | ✅ | ✅ |

---

## Go-Live Checklist

Use this checklist to verify readiness before deploying to production.

### ✅ Must-pass before go-live

- [ ] `go build ./...` passes with zero errors
- [ ] `go test -race ./...` — all unit tests pass
- [ ] `go vet ./...` — zero vet warnings
- [ ] `GET /healthz` returns HTTP 200 on all three services
- [ ] `GET /metrics` returns valid Prometheus text format on all three services
- [ ] Docker images build successfully for all three services:
  ```bash
  docker build -f docker/api/Dockerfile -t task-scheduler-api:v1.0.0 .
  docker build -f docker/scheduler/Dockerfile -t task-scheduler-scheduler:v1.0.0 .
  docker build -f docker/worker/Dockerfile -t task-scheduler-worker:v1.0.0 .
  ```
- [ ] `docker compose up --build` starts successfully; `curl http://localhost:8080/healthz` returns `{"service":"task-scheduler-api","status":"ok"}`
- [ ] Database migrations applied on target database: `migrate -path db/migrations -database "$DATABASE_URL" up`
- [ ] `DATABASE_URL` Secret in `k8s/configmap.yaml` replaced with real credentials (or external secret reference)
- [ ] K8s manifests applied in order: `namespace.yaml` → `configmap.yaml` → all three Deployments
- [ ] All three Deployment rollouts complete: `kubectl -n task-scheduler rollout status deployment/api deployment/scheduler deployment/worker`
- [ ] Liveness and readiness probes healthy: `kubectl -n task-scheduler get pods`
- [ ] Prometheus successfully scraping all three targets (check `Status → Targets` in Prometheus UI)
- [ ] CI pipeline (`.github/workflows/ci.yaml`) is green on the release commit

### ⚠️ Known limitations (address post-launch)

- **In-memory queue**: The scheduler uses an in-memory FIFO queue. Restarting the scheduler pod will drop any queued-but-not-yet-dequeued tasks. Integrate Redis Streams or a PostgreSQL-backed queue before enabling high-availability deployments.
- **Scheduler: no cron triggering**: `cmd/scheduler/main.go` initialises the scheduler but does not yet parse `ScheduleCron` fields and trigger `WorkflowRun` creation on schedule. A cron library (e.g. `github.com/robfig/cron/v3`) and a trigger loop need to be added.
- **No DAG dependency enforcement**: `TaskDependency` rows are persisted in the database but the worker does not check upstream task completion before executing downstream tasks.
- **API is unauthenticated**: Add JWT/API-key middleware before exposing the API to the public internet.
- **No end-to-end tests**: Unit tests cover all layers in isolation; an integration smoke test (`docker compose up` → create workflow → verify `success` status) would close the last gap.
