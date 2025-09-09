package httpx

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"search-engine/internal/index"
	"search-engine/internal/search"
	"search-engine/internal/snippet"

	"github.com/go-chi/chi/v5"
)

type appLike interface {
	// expose what handlers need (minimal interface)
}

type App struct {
	DB  *sql.DB
	Idx *index.Engine
}

func RegisterRoutes(r *chi.Mux, db *sql.DB, idx *index.Engine) {
	wrapped := &App{DB: db, Idx: idx}
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { /* ... */ })
	r.Get("/search", wrapped.search)
	r.Post("/admin/reindex", wrapped.reindex)
}

func (a *App) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q required", http.StatusBadRequest)
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	ps, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if ps < 1 {
		ps = 10
	}

	terms, items := search.Query(a.Idx, q, page*ps)
	total := len(items)
	start := (page - 1) * ps
	if start > total {
		start = total
	}
	end := start + ps
	if end > total {
		end = total
	}
	items = items[start:end]

	// fill snippets
	for i := range items {
		items[i].Snippet = snippet.Make(r.Context(), a.DB, items[i].DocID, terms, 200)
	}

	resp := struct {
		Total    int           `json:"total"`
		Page     int           `json:"page"`
		PageSize int           `json:"page_size"`
		Results  []search.Item `json:"results"`
	}{Total: total, Page: page, PageSize: ps, Results: items}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (a *App) reindex(w http.ResponseWriter, r *http.Request) {
	ix, err := index.Build(a.DB)
	if err != nil {
		http.Error(w, "reindex failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	a.Idx = ix
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("reindexed"))
}
