package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"org-structure-api/config"
	"org-structure-api/internal/handler"
	"org-structure-api/internal/repository"
	"org-structure-api/internal/usecase"
)

func main() {
	log.Println("Starting organizational structure API...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: cmd/api/main.go: %v", err)
	}
	log.Println("Configuration successfully loaded")

	db, err := repository.NewPostgresDB(cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: cmd/api/main.go: %v", err)
	}
	log.Println("Database successfully connected")

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying SQL DB: cmd/api/main.go: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Database ping failed: cmd/api/main.go: %v", err)
	}
	log.Println("Database ping successful")

	deptRepo := repository.NewDepartmentRepository(db)
	empRepo := repository.NewEmployeeRepository(db)

	deptUseCase := usecase.NewDepartmentUseCase(deptRepo)
	empUseCase := usecase.NewEmployeeUseCase(empRepo, deptRepo)

	deptHandler := handler.NewDepartmentHandler(deptUseCase)
	empHandler := handler.NewEmployeeHandler(empUseCase)
	
	mux := http.NewServeMux()

	mux.HandleFunc("POST /departments/", deptHandler.Create)
	mux.HandleFunc("GET /departments/{id}", deptHandler.GetByID)
	mux.HandleFunc("PATCH /departments/{id}", deptHandler.Update)
	mux.HandleFunc("DELETE /departments/{id}", deptHandler.Delete)

	mux.HandleFunc("POST /departments/{id}/employees/", empHandler.Create)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"OK","message":"Server is running"}`))
	})

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server is listening on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server failed: cmd/api/main.go: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: cmd/api/main.go: %v", err)
	}

	log.Println("Server stopped dynamic resources. Goodbye!")
}
