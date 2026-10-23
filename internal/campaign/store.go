package campaign

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

// Campaign mirrors a row in the campaigns table.
// We keep money as int64 micros, matching the schema.
type Campaign struct {
    ID             int64
    Name           string
    BidMicros      int64
    BudgetMicros   int64
    SpentMicros    int64
    TargetCategory string
    Active         bool
}

// Store wraps the database connection pool and exposes
// the queries the rest of the app needs.
type Store struct {
    pool *pgxpool.Pool
}

// NewStore creates a Store from an existing pool.
func NewStore(pool *pgxpool.Pool) *Store {
    return &Store{pool: pool}
}

// EligibleForCategory returns all active campaigns targeting the given
// category that still have budget remaining. This is the exact query the
// auction hot path runs to find competitors for a bid request.
func (s *Store) EligibleForCategory(ctx context.Context, category string) ([]Campaign, error) {
    const query = `
        select id, name, bid_micros, budget_micros, spent_micros, target_category, active
        from campaigns
        where target_category = $1
          and active = true
          and spent_micros < budget_micros
        order by bid_micros desc
    `

    rows, err := s.pool.Query(ctx, query, category)
    if err != nil {
        return nil, fmt.Errorf("query eligible campaigns: %w", err)
    }
    defer rows.Close()

    var campaigns []Campaign
    for rows.Next() {
        var c Campaign
        if err := rows.Scan(
            &c.ID, &c.Name, &c.BidMicros, &c.BudgetMicros,
            &c.SpentMicros, &c.TargetCategory, &c.Active,
        ); err != nil {
            return nil, fmt.Errorf("scan campaign row: %w", err)
        }
        campaigns = append(campaigns, c)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate campaign rows: %w", err)
    }

    return campaigns, nil
}
