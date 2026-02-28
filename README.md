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
- [ ] **Phase 7** — Persistence layer (PostgreSQL — wire up repositories to real DB)
- [ ] **Phase 8** — Observability (metrics, structured logging, tracing)
