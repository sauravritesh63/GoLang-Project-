// Package handler provides the HTTP handler layer for the distributed task
// scheduler API. Each handler delegates to the service layer and writes a
// JSON response.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/api/service"
	ws "github.com/sauravritesh63/GoLang-Project-/internal/api/websocket"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
)

// Handler groups the service and WebSocket hub dependencies for all HTTP
// handlers. Create one via New and register routes via RegisterRoutes.
type Handler struct {
	svc *service.Service
	hub *ws.Hub
}

// New constructs a Handler with the supplied service and WebSocket hub.
func New(svc *service.Service, hub *ws.Hub) *Handler {
	return &Handler{svc: svc, hub: hub}
}

// RegisterRoutes mounts all API routes onto the supplied Gin engine.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/workflows", h.createWorkflow)
	r.GET("/workflows", h.listWorkflows)
	r.POST("/workflows/:id/trigger", h.triggerWorkflow)
	r.GET("/workflow-runs", h.listWorkflowRuns)
	r.GET("/task-runs", h.listTaskRuns)
	r.GET("/workers", h.listWorkers)
	r.GET("/ws/updates", h.serveWS)
}

// createWorkflow handles POST /workflows.
func (h *Handler) createWorkflow(c *gin.Context) {
	var in service.CreateWorkflowInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wf, err := h.svc.CreateWorkflow(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, wf)
}

// listWorkflows handles GET /workflows with optional ?offset=&limit= pagination.
func (h *Handler) listWorkflows(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	wfs, err := h.svc.ListWorkflows(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wfs)
}

// triggerWorkflow handles POST /workflows/{id}/trigger.
func (h *Handler) triggerWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow id"})
		return
	}
	run, err := h.svc.TriggerWorkflow(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Broadcast the new workflow run event to connected WebSocket clients.
	h.hub.Broadcast(c.Request.Context(), ws.Event{
		Type:    ws.EventWorkflowStatus,
		Payload: run,
	})
	c.JSON(http.StatusCreated, run)
}

// listWorkflowRuns handles GET /workflow-runs with optional ?status= filter.
func (h *Handler) listWorkflowRuns(c *gin.Context) {
	status := domain.Status(c.Query("status"))
	runs, err := h.svc.ListWorkflowRuns(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, runs)
}

// listTaskRuns handles GET /task-runs with optional ?status= filter.
func (h *Handler) listTaskRuns(c *gin.Context) {
	status := domain.Status(c.Query("status"))
	trs, err := h.svc.ListTaskRuns(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trs)
}

// listWorkers handles GET /workers.
func (h *Handler) listWorkers(c *gin.Context) {
	workers, err := h.svc.ListWorkers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, workers)
}

// serveWS upgrades the connection to WebSocket and streams real-time events.
func (h *Handler) serveWS(c *gin.Context) {
	h.hub.ServeWS(c.Writer, c.Request)
}
