// Package api wires up the Gin engine with all routes and middleware for the
// distributed task scheduler REST API.
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sauravritesh63/GoLang-Project-/internal/api/handler"
	"github.com/sauravritesh63/GoLang-Project-/internal/api/service"
	ws "github.com/sauravritesh63/GoLang-Project-/internal/api/websocket"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
)

// NewRouter constructs and returns a configured *gin.Engine.
// All dependencies are injected via the repository interfaces so that the
// router can be used in tests with mock implementations.
func NewRouter(
	workflows repository.WorkflowRepository,
	workflowRuns repository.WorkflowRunRepository,
	taskRuns repository.TaskRunRepository,
	workers repository.WorkerRepository,
) *gin.Engine {
	svc := service.New(workflows, workflowRuns, taskRuns, workers)
	hub := ws.NewHub()
	h := handler.New(svc, hub)

	r := gin.New()
	r.Use(gin.Recovery())
	h.RegisterRoutes(r)

	// Expose Prometheus metrics at /metrics.
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}
