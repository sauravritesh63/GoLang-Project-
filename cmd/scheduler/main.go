// Package main is the entry point for the distributed task scheduler service.
// It wires up the in-memory queue, task and worker repositories, the cron
// trigger, and waits for shutdown. In production, replace the in-memory stores
// with Redis / Postgres backed implementations via environment variables.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sauravritesh63/GoLang-Project-/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository/mock"
	"github.com/sauravritesh63/GoLang-Project-/observability/metrics"
	"github.com/sauravritesh63/GoLang-Project-/scheduler"
)

func main() {
	metricsPort := getEnv("METRICS_PORT", "9090")

	// Register Prometheus metrics for this scheduler process. The Collector is
	// not stored because promauto registers all metrics with the default registry
	// on construction; the /metrics handler will serve them automatically.
	_ = metrics.New()

	// Expose /metrics and /healthz on a dedicated port.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"task-scheduler-scheduler"}`))
	})
	go func() {
		log.Printf("Scheduler metrics server listening on :%s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil && err != http.ErrServerClosed {
			log.Printf("metrics server error: %v", err)
		}
	}()

	queue := scheduler.NewMemQueue()
	taskRepo := newMemTaskRepo()
	workerRepo := newMemWorkerRepo()

	// In-memory workflow and workflow-run repositories (replace with Postgres in
	// production). The CronTrigger reads active workflows at startup and
	// schedules WorkflowRun creation according to each workflow's ScheduleCron
	// field.
	wfRepo := mock.NewWorkflowRepo()
	wfRunRepo := mock.NewWorkflowRunRepo()

	// Scheduler — validates and enqueues tasks.
	sched := scheduler.New(taskRepo, workerRepo, queue)
	log.Printf("Scheduler initialised (queue depth: %T)", sched)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// CronTrigger — creates WorkflowRuns on schedule.
	ct := scheduler.NewCronTrigger(wfRepo, wfRunRepo)
	if err := ct.Start(ctx); err != nil {
		log.Printf("CronTrigger: failed to start: %v", err)
	}
	defer ct.Stop()

	log.Println("Scheduler service started; waiting for shutdown signal")
	<-ctx.Done()
	log.Println("Scheduler service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ── in-memory stores (replace with Redis/Postgres in production) ──────────────

type memTaskRepo struct {
	mu    sync.RWMutex
	store map[string]*domain.Task
}

func newMemTaskRepo() *memTaskRepo {
	return &memTaskRepo{store: make(map[string]*domain.Task)}
}

func (r *memTaskRepo) Save(_ context.Context, t *domain.Task) error {
	r.mu.Lock()
	cp := *t
	r.store[t.ID] = &cp
	r.mu.Unlock()
	return nil
}

func (r *memTaskRepo) FindByID(_ context.Context, id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.store[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *memTaskRepo) FindByStatus(_ context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Task
	for _, t := range r.store {
		if t.Status == status {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memTaskRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return domain.ErrTaskNotFound
	}
	delete(r.store, id)
	return nil
}

type memWorkerRepo struct {
	mu    sync.RWMutex
	store map[string]*domain.Worker
}

func newMemWorkerRepo() *memWorkerRepo {
	return &memWorkerRepo{store: make(map[string]*domain.Worker)}
}

func (r *memWorkerRepo) Save(_ context.Context, w *domain.Worker) error {
	r.mu.Lock()
	cp := *w
	r.store[w.ID] = &cp
	r.mu.Unlock()
	return nil
}

func (r *memWorkerRepo) FindByID(_ context.Context, id string) (*domain.Worker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.store[id]
	if !ok {
		return nil, domain.ErrWorkerNotFound
	}
	cp := *w
	return &cp, nil
}

func (r *memWorkerRepo) FindAvailable(_ context.Context) ([]*domain.Worker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Worker
	for _, w := range r.store {
		if w.HasCapacity() {
			cp := *w
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memWorkerRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return domain.ErrWorkerNotFound
	}
	delete(r.store, id)
	return nil
}
