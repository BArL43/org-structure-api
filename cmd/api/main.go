package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	if err := run(); err != nil {
		log.Fatalf("organizational structure API stopped: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	db, err := repository.NewPostgresDB(cfg.GetDSN())
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database ping: %w", err)
	}

	deptRepo := repository.NewDepartmentRepository(db)
	empRepo := repository.NewEmployeeRepository(db)
	deptHandler := handler.NewDepartmentHandler(usecase.NewDepartmentUseCase(deptRepo))
	empHandler := handler.NewEmployeeHandler(usecase.NewEmployeeUseCase(empRepo, deptRepo))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /departments/", deptHandler.Create)
	mux.HandleFunc("GET /departments/{id}", deptHandler.GetByID)
	mux.HandleFunc("PATCH /departments/{id}", deptHandler.Update)
	mux.HandleFunc("DELETE /departments/{id}", deptHandler.Delete)
	mux.HandleFunc("POST /departments/{id}/employees/", empHandler.Create)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			writeHealth(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		writeHealth(w, http.StatusOK, "ok")
	})

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server listening on :%s", cfg.ServerPort)
		serverErr <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalCtx.Done():
		log.Println("shutdown signal received")
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	return nil
}

func writeHealth(w http.ResponseWriter, status int, state string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": state})
}
