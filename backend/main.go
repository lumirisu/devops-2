package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	dsn := "host=" + env("DB_HOST", "db") +
		" port=" + env("DB_PORT", "5432") +
		" user=" + env("DB_USER", "app") +
		" password=" + env("DB_PASSWORD", "secret") +
		" dbname=" + env("DB_NAME", "appdb") +
		" sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := waitForDatabase(db, 30, 2*time.Second); err != nil {
		log.Fatal(err)
	}

	staticDir := env("STATIC_DIR", "/app/static")
	fileServer := http.FileServer(http.Dir(staticDir))

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.QueryContext(r.Context(), `SELECT id, name FROM items ORDER BY id`)
		if err != nil {
			http.Error(w, "failed to query items", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]Item, 0)
		for rows.Next() {
			var item Item
			if err := rows.Scan(&item.ID, &item.Name); err != nil {
				http.Error(w, "failed to read item", http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "failed to iterate items", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(items); err != nil {
			log.Printf("encode items response: %v", err)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	addr := ":" + env("HTTP_PORT", "8080")
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func waitForDatabase(db *sql.DB, attempts int, delay time.Duration) error {
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := db.Ping(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		log.Printf("database is not ready (attempt %d/%d): %v", attempt, attempts, lastErr)
		time.Sleep(delay)
	}
	return lastErr
}

func env(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return def
}
