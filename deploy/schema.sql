-- campaigns holds one row per advertiser campaign.
-- This is the durable source of truth for the auction.
create table if not exists campaigns (
    id              bigserial primary key,
    name            text        not null,
    -- bid amount in micro-dollars (1 dollar = 1,000,000 micros).
    -- ad-tech uses integer micros instead of floats to avoid
    -- rounding errors when doing money math at high volume.
    bid_micros      bigint      not null,
    -- total budget the campaign is allowed to spend, in micros.
    budget_micros   bigint      not null,
    -- how much has been spent so far, in micros. starts at 0.
    spent_micros    bigint      not null default 0,
    -- simple targeting: which ad category this campaign wants.
    -- we start simple and add richer targeting later.
    target_category text        not null,
    -- whether the campaign is currently eligible to bid.
    active          boolean     not null default true,
    created_at      timestamptz not null default now()
);

-- index to quickly fetch active campaigns for a given category
-- during an auction. the auction hot path queries by these two columns.
create index if not exists idx_campaigns_category_active
    on campaigns (target_category, active);
