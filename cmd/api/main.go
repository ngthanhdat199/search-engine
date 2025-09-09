package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"search-engine/internal/httpx"
	"search-engine/internal/index"
	"search-engine/internal/store"

	"github.com/go-chi/chi/v5"
)

type App struct {
	DB  *sql.DB
	Idx *index.Engine
	Log *log.Logger
}

func main() {
	logger := log.New(os.Stdout, "[search-engine] ", log.LstdFlags|log.Lshortfile)

	// DB (SQLite file path can be overridden by env)
	dbPath := getenv("SQLITE_PATH", "mini_search.db")
	db, err := store.OpenSQLite(dbPath)
	if err != nil {
		logger.Fatalf("open db: %v", err)
	}
	if err := store.RunMigrations(db, "migrations/0001_init.sql"); err != nil {
		logger.Fatalf("migrate: %v", err)
	}

	// Build index in RAM
	ix, err := index.Build(db)
	if err != nil {
		logger.Fatalf("index build: %v", err)
	}
	logger.Printf("index ready: docs=%d, terms=%d", ix.N, len(ix.Postings))

	// app := &App{DB: db, Idx: ix, Log: logger}

	// Router
	r := chi.NewRouter()
	httpx.RegisterRoutes(r, db, ix)

	port := atoi(getenv("PORT", "8000"))
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logger.Printf("listening on http://%s ...", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatalf("server: %v", err)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func atoi(s string) int { i, _ := strconv.Atoi(s); return i }
