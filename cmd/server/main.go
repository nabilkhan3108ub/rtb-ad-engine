package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/nabilkhan3108ub/adserving/internal/campaign"
)

type healthResponse struct {
    Status string `json:"status"`
    Time   string `json:"time"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    resp := healthResponse{
        Status: "ok",
        Time:   time.Now().UTC().Format(time.RFC3339),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

// campaignsHandler returns eligible campaigns for a given category.
// Example: GET /campaigns?category=sports
func campaignsHandler(store *campaign.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        category := r.URL.Query().Get("category")
        if category == "" {
            http.Error(w, "missing 'category' query param", http.StatusBadRequest)
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
        defer cancel()

        campaigns, err := store.EligibleForCategory(ctx, category)
        if err != nil {
            log.Printf("error fetching campaigns: %v", err)
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(campaigns)
    }
}

func main() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://adserving:adserving@localhost:5433/adserving"
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        log.Fatalf("unable to create connection pool: %v", err)
    }
    defer pool.Close()

    if err := pool.Ping(ctx); err != nil {
        log.Fatalf("unable to reach database: %v", err)
    }
    log.Printf("connected to postgres")

    store := campaign.NewStore(pool)

    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("/campaigns", campaignsHandler(store))

    addr := ":8080"
    log.Printf("ad-serving engine starting on %s", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
