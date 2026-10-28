package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/nabilkhan3108ub/adserving/internal/auction"
    "github.com/nabilkhan3108ub/adserving/internal/campaign"
)

type healthResponse struct {
    Status string `json:"status"`
    Time   string `json:"time"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    resp := healthResponse{Status: "ok", Time: time.Now().UTC().Format(time.RFC3339)}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

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

// bidRequest is the incoming ad opportunity. category selects which
// campaigns compete; userID is carried for future frequency capping.
type bidRequest struct {
    Category string `json:"category"`
    UserID   string `json:"user_id"`
}

// bidResponse is what we return to the caller: the auction outcome.
type bidResponse struct {
    Filled              bool   `json:"filled"`
    WinningCampaignID   int64  `json:"winning_campaign_id,omitempty"`
    WinningCampaignName string `json:"winning_campaign_name,omitempty"`
    ClearingPriceMicros int64  `json:"clearing_price_micros,omitempty"`
    NumBidders          int    `json:"num_bidders"`
}

// bidHandler is the core ad-exchange endpoint. It accepts a bid request,
// pulls eligible campaigns, runs a second-price auction, and returns the
// winning ad with its clearing price.
func bidHandler(store *campaign.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
            return
        }

        var req bidRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid JSON body", http.StatusBadRequest)
            return
        }
        if req.Category == "" {
            http.Error(w, "missing 'category' in body", http.StatusBadRequest)
            return
        }

        // hard latency budget on the whole auction path.
        ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
        defer cancel()

        candidates, err := store.EligibleForCategory(ctx, req.Category)
        if err != nil {
            log.Printf("bid: error fetching campaigns: %v", err)
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }

        result := auction.RunSecondPrice(candidates)

        resp := bidResponse{NumBidders: result.NumBidders}
        if result.Winner != nil {
            resp.Filled = true
            resp.WinningCampaignID = result.Winner.ID
            resp.WinningCampaignName = result.Winner.Name
            resp.ClearingPriceMicros = result.ClearingPriceMicros
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
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
    mux.HandleFunc("/bid", bidHandler(store))

    addr := ":8080"
    log.Printf("ad-serving engine starting on %s", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
