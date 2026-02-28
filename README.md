# Mini Airflow — Distributed Task Scheduler

A lightweight, distributed task scheduler inspired by Apache Airflow, built with Go and clean architecture principles.

---

## Architecture (placeholder)

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

| Field          | Type        | JSON key        | Description                          |
|----------------|-------------|-----------------|--------------------------------------|
| `ID`           | `uuid.UUID` | `id`            | Unique workflow identifier           |
| `Name`         | `string`    | `name`          | Human-readable workflow name         |
| `Description`  | `string`    | `description`   | Optional description                 |
| `ScheduleCron` | `string`    | `schedule_cron` | Cron expression for scheduling       |
| `IsActive`     | `bool`      | `is_active`     | Whether the workflow is enabled      |
| `CreatedAt`    | `time.Time` | `created_at`    | Creation timestamp                   |

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

| Field             | Type        | JSON key             | Description                          |
|-------------------|-------------|----------------------|--------------------------------------|
| `ID`              | `uuid.UUID` | `id`                 | Unique dependency identifier         |
| `TaskID`          | `uuid.UUID` | `task_id`            | The downstream (waiting) task        |
| `DependsOnTaskID` | `uuid.UUID` | `depends_on_task_id` | The upstream (prerequisite) task     |

#### `WorkflowRun`
A single execution instance of a `Workflow`.

| Field        | Type         | JSON key      | Description                          |
|--------------|--------------|---------------|--------------------------------------|
| `ID`         | `uuid.UUID`  | `id`          | Unique run identifier                |
| `WorkflowID` | `uuid.UUID`  | `workflow_id` | The workflow being executed          |
| `Status`     | `Status`     | `status`      | Current lifecycle status             |
| `StartedAt`  | `time.Time`  | `started_at`  | When the run began                   |
| `FinishedAt` | `*time.Time` | `finished_at` | When the run completed (nullable)    |

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

| Field           | Type           | JSON key         | Description                          |
|-----------------|----------------|------------------|--------------------------------------|
| `ID`            | `uuid.UUID`    | `id`             | Unique worker identifier             |
| `Hostname`      | `string`       | `hostname`       | Network hostname of the worker       |
| `LastHeartbeat` | `time.Time`    | `last_heartbeat` | Most recent heartbeat timestamp      |
| `Status`        | `WorkerStatus` | `status`         | `active` or `inactive`               |

---

## Test Structure (`internal/domain/domain_test.go`)

All tests use the standard `testing` package with no external test frameworks.

| Test name                          | What it covers                                                   |
|------------------------------------|------------------------------------------------------------------|
| `TestStatusConstants`              | Status enum string values (`pending`, `running`, `success`, `failed`) |
| `TestWorkerStatusConstants`        | WorkerStatus enum string values (`active`, `inactive`)           |
| `TestWorkflowInstantiation`        | Struct field assignment for Workflow                             |
| `TestWorkflowJSONRoundtrip`        | JSON marshal/unmarshal round-trip for Workflow                   |
| `TestTaskInstantiation`            | Struct field assignment for Task                                 |
| `TestTaskJSONRoundtrip`            | JSON marshal/unmarshal round-trip for Task                       |
| `TestTaskDependencyInstantiation`  | Struct field assignment for TaskDependency                       |
| `TestTaskDependencyJSONRoundtrip`  | JSON marshal/unmarshal round-trip for TaskDependency             |
| `TestWorkflowRunInstantiation`     | Struct field assignment; nil FinishedAt for running run          |
| `TestWorkflowRunOptionalFinishedAt`| `finished_at` omitted from JSON when nil                        |
| `TestWorkflowRunWithFinishedAt`    | `finished_at` preserved in JSON when set                         |
| `TestTaskRunInstantiation`         | Struct field assignment for TaskRun                              |
| `TestTaskRunOptionalFinishedAt`    | `finished_at` omitted from JSON when nil                        |
| `TestWorkerInstantiation`          | Struct field assignment for Worker                               |
| `TestWorkerJSONRoundtrip`          | JSON marshal/unmarshal round-trip for Worker                     |
| `TestStatusJSONField`              | Status enum serialises to correct JSON string value              |

Run all tests:

```bash
go test ./...
```

---

## Project Roadmap

- [x] **Phase 1** — Domain models (`internal/domain`)
  - Core structs: `Workflow`, `Task`, `TaskDependency`, `WorkflowRun`, `TaskRun`, `Worker`
  - Typed `Status` and `WorkerStatus` enums
  - Strict JSON tags on all fields
  - 16 unit tests: struct instantiation, status enums, JSON round-trips, nullable field handling
- [ ] **Phase 2** — Repository interfaces & in-memory implementations
- [ ] **Phase 3** — Scheduler service (cron-based workflow triggering)
- [ ] **Phase 4** — Worker service (task execution, heartbeat, retry logic)
- [ ] **Phase 5** — REST API (workflow/run management endpoints)
- [ ] **Phase 6** — Persistence layer (PostgreSQL or SQLite)
- [ ] **Phase 7** — Observability (metrics, structured logging, tracing)

---

## Phase 1 — Domain Models

Phase 1 establishes the **foundation** of the project: pure Go data models with no infrastructure dependencies.

**What is included:**
- `internal/domain/domain.go` — all domain structs and typed const enums
- `internal/domain/domain_test.go` — 16 unit tests covering struct instantiation, status enum values, JSON serialisation round-trips, and nullable field (`omitempty`) behaviour

**Design decisions:**
- **UUIDs** (`github.com/google/uuid`) are used for all IDs to avoid collision in a distributed environment.
- **`time.Time`** is used for all timestamps; nullable timestamps (`FinishedAt`) use `*time.Time` so they are omitted from JSON when unset (`omitempty`).
- **Typed `Status` constants** (not plain strings) prevent invalid states from being passed around the codebase.
- No database, repository, or service code is introduced in this phase — clean architecture keeps the domain layer dependency-free from infrastructure.

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
