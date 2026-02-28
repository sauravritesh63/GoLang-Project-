# Distributed Task Scheduler — Go

A distributed task scheduler built in Go, following a clean-architecture approach.

---

## Current State (Phase 1 — Domain Layer)

### What exists

| Item | Path | Notes |
|---|---|---|
| Go module | `go.mod` | `github.com/sauravritesh63/distributed-task-scheduler` |
| Domain entities | `domain/task.go` | `Task` struct, status constants, priority levels, `Validate()`, `CanRetry()`, `IsTerminal()` |
| Domain entities | `domain/worker.go` | `Worker` struct, status constants, `Validate()`, `HasCapacity()`, `IsAlive()` |
| Domain interfaces | `domain/interfaces.go` | `TaskRepository`, `WorkerRepository`, `Queue`, `Scheduler` |
| Sentinel errors | `domain/errors.go` | `ErrTaskNotFound`, `ErrWorkerNotFound`, `ErrQueueEmpty`, `ErrTaskInvalid`, `ErrWorkerInvalid` |
| Unit tests | `domain/domain_test.go` | 18 tests — all passing |

### Compilation

```
go build ./...   # ✅ compiles cleanly (Go 1.24)
go test ./...    # ✅ 18/18 tests pass
```

---

## Architecture

```
┌─────────────────────────────────────────────┐
│                  domain/                    │  ← Phase 1 ✅
│  Task · Worker · Queue · Scheduler          │
│  TaskRepository · WorkerRepository          │
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

## Phases

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Domain layer — entities, interfaces, sentinel errors, tests | ✅ Done |
| 2 | Service & storage layer — in-memory `Queue`, `TaskRepository`, `WorkerRepository`; `Scheduler` use-case | 🔲 Next |
| 3 | HTTP API — REST endpoints for submitting, cancelling, and querying tasks | 🔲 Planned |
| 4 | Worker executor — goroutine pool that dequeues and runs tasks | 🔲 Planned |
| 5 | Observability — structured logging, Prometheus metrics, health endpoint | 🔲 Planned |

---

## Next Recommended Phase (Phase 2 — Service & Storage Layer)

### Goal
Wire the domain interfaces to concrete, in-memory implementations so the
scheduler can run end-to-end without an external dependency.

### Suggested implementation tasks

1. **`storage/memory/task_repo.go`** — thread-safe, in-memory `TaskRepository`
   backed by a `sync.RWMutex`-protected map; `FindByStatus` sorts by priority
   then `ScheduledAt`.

2. **`storage/memory/worker_repo.go`** — thread-safe, in-memory
   `WorkerRepository`; `FindAvailable` filters by `HasCapacity()`.

3. **`storage/memory/queue.go`** — priority-queue implementation of `Queue`
   using `container/heap`; `Dequeue` blocks via a `sync.Cond`.

4. **`scheduler/service.go`** — `SchedulerService` struct that implements the
   `domain.Scheduler` interface; `Submit` validates, persists, and enqueues the
   task; `Cancel` sets status to `failed`; `Status` delegates to the repository.

5. **Tests** — table-driven tests for each of the above, exercising concurrency
   with `t.Parallel()` and the race detector (`go test -race ./...`).

### Running the project

```bash
# build
go build ./...

# test (with race detector)
go test -race ./...
```