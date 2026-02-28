// Package main is the entry point for the distributed task scheduler API server.
// It reads configuration from environment variables, connects to PostgreSQL,
// and serves the HTTP API using Gin.
package main

import (
	"log"
	"os"

	"github.com/sauravritesh63/GoLang-Project-/internal/api"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository/mock"
	pgRepo "github.com/sauravritesh63/GoLang-Project-/internal/repository/postgres"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	port := getEnv("PORT", "8080")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		db, err := gorm.Open(pgdriver.Open(dbURL), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}
		r := api.NewRouter(
			pgRepo.NewWorkflowRepo(db),
			pgRepo.NewWorkflowRunRepo(db),
			pgRepo.NewTaskRunRepo(db),
			pgRepo.NewWorkerRepo(db),
		)
		log.Printf("API server listening on :%s (postgres)", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	} else {
		log.Println("DATABASE_URL not set — using in-memory repositories")
		r := api.NewRouter(
			mock.NewWorkflowRepo(),
			mock.NewWorkflowRunRepo(),
			mock.NewTaskRunRepo(),
			mock.NewWorkerRepo(),
		)
		log.Printf("API server listening on :%s (in-memory)", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
