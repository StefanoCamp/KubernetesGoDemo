package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type createItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type apiError struct {
	Error string `json:"error"`
}

type application struct {
	mu        sync.RWMutex
	items     []item
	nextID    int
	startedAt time.Time
	name      string
	env       string
}

func main() {
	app := &application{
		items: []item{
			{
				ID:          1,
				Name:        "First demo item",
				Description: "Seed item returned by the API",
				CreatedAt:   time.Now().UTC(),
			},
		},
		nextID:    2,
		startedAt: time.Now().UTC(),
		name:      getEnv("APP_NAME", "kubernetes-go-api-demo"),
		env:       getEnv("APP_ENV", "local"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.healthHandler)
	mux.HandleFunc("/api/items", app.itemsHandler)

	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           requestLogger(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting %s on port %s", app.name, port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "ok",
		"service":        app.name,
		"environment":    app.env,
		"uptime_seconds": int(time.Since(app.startedAt).Seconds()),
	})
}

func (app *application) itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.listItems(w, r)
	case http.MethodPost:
		app.createItem(w, r)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (app *application) listItems(w http.ResponseWriter, _ *http.Request) {
	app.mu.RLock()
	items := make([]item, len(app.items))
	copy(items, app.items)
	app.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"count": len(items),
	})
}

func (app *application) createItem(w http.ResponseWriter, r *http.Request) {
	var input createItemRequest

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		message := "invalid JSON request body"
		if errors.Is(err, io.EOF) {
			message = "request body is required"
		}
		writeJSON(w, http.StatusBadRequest, apiError{Error: message})
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "name is required"})
		return
	}

	app.mu.Lock()
	newItem := item{
		ID:          app.nextID,
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   time.Now().UTC(),
	}
	app.items = append(app.items, newItem)
	app.nextID++
	app.mu.Unlock()

	writeJSON(w, http.StatusCreated, newItem)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func methodNotAllowed(w http.ResponseWriter, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	writeJSON(w, http.StatusMethodNotAllowed, apiError{
		Error: "method not allowed",
	})
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, strconv.FormatInt(time.Since(start).Milliseconds(), 10)+"ms")
	})
}
