package postgres_test

import (
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository/postgres"
)

// Compile-time checks that each postgres repo satisfies the corresponding
// repository interface.
var (
	_ repository.WorkflowRepository    = (*postgres.WorkflowRepo)(nil)
	_ repository.TaskRepository        = (*postgres.TaskRepo)(nil)
	_ repository.WorkflowRunRepository = (*postgres.WorkflowRunRepo)(nil)
	_ repository.TaskRunRepository     = (*postgres.TaskRunRepo)(nil)
	_ repository.WorkerRepository      = (*postgres.WorkerRepo)(nil)
)
