package auction

import "github.com/nabilkhan3108ub/adserving/internal/campaign"

// Result describes the outcome of an auction.
type Result struct {
	Winner              *campaign.Campaign `json:"winner"`
	ClearingPriceMicros int64              `json:"clearing_price_micros"`
	NumBidders          int                `json:"num_bidders"`
}

// RunSecondPrice runs a second-price (Vickrey) auction over the given
// candidates. The highest bidder wins but pays just above the second
// highest bid. This incentivizes advertisers to bid their true value.
func RunSecondPrice(candidates []campaign.Campaign) Result {
	if len(candidates) == 0 {
		return Result{Winner: nil, ClearingPriceMicros: 0, NumBidders: 0}
	}

	var highest, second *campaign.Campaign
	for i := range candidates {
		c := &candidates[i]
		if highest == nil || c.BidMicros > highest.BidMicros {
			second = highest
			highest = c
		} else if second == nil || c.BidMicros > second.BidMicros {
			second = c
		}
	}

	var clearing int64
	if second != nil {
		clearing = second.BidMicros + 1
	} else {
		clearing = highest.BidMicros
	}

	return Result{
		Winner:              highest,
		ClearingPriceMicros: clearing,
		NumBidders:          len(candidates),
	}
}