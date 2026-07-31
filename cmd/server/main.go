package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	db "github.com/P4rz1val22/outreach-tool/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readyHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func main() {
	// Establish context and connection pooling
	ctx := context.Background()
	godotenv.Load()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Connection error: %v", err)
	}

	queries := db.New(pool)
	someTestUUID := uuid.New()
	contacts, err := queries.ListContacts(ctx, someTestUUID)
	if err != nil {
		log.Printf("ListContacts error: %v", err)
	} else {
		log.Printf("ListContacts returned %d rows", len(contacts))
	}

	// Register handlers
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler(pool))

	// Start server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{Addr: ":" + os.Getenv("PORT")}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received, starting graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("shutting down server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	} else {
		log.Println("server shut down cleanly")
	}

	log.Println("closing db pool...")
	pool.Close()
	log.Println("db pool closed, exiting")
}
