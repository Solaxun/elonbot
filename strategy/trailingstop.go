package strategy

import "math"

type trailingStopMarket struct {
	stopprice   float64
	trailingpct float64
	qtyheld     int
	tradeon     bool
	maxprice    float64
}

// NewTrailingStop creates a strategy with a stop loss initially set at a given percentage below the current price.
// as price moves up, the trailing stop will be adjusted to maintain the percentage distance specified below the price.
// as price falls, the stop remains in place (otherwise the stop would never be hit).  Once the stop is hit, the position
// is sold at either limit or market, and the strategy is turned off.
func NewTrailingStopMarket(price, trailingpct, maxprice float64, qtyheld int, tradeon bool) *trailingStopMarket {
	return &trailingStopMarket{
		stopprice:   price * (1 - trailingpct),
		trailingpct: trailingpct,
		qtyheld:     qtyheld,
		tradeon:     tradeon,
		maxprice:    maxprice,
	}
}

type trailStopMarketArgs struct {
	stopprice   float64
	trailingAmt float64
	trailPct    float64
	tradeon     bool
}

func (tsm *trailingStopMarket) Update(price float64) (*trailingStopMarket, bool) {
	if price <= tsm.stopprice {
		return tsm, true
	}
	tsm.stopprice = math.Max(tsm.stopprice, price)
	return tsm, false
}

type trailingStopLimit struct {
	stopprice   float64
	trailingAmt float64
	limitOffset float64
	qtyheld     int
	tradeon     bool
	maxprice    float64
}

// NewTrailingStop creates a strategy with a stop loss initially set at a given percentage below the current price.
// as price moves up, the trailing stop will be adjusted to maintain the percentage distance specified below the price.
// as price falls, the stop remains in place (otherwise the stop would never be hit).  Once the stop is hit, the position
// is sold at either limit or market, and the strategy is turned off.
func NewTrailingStopLimit(price, trailingAmt, limitOffset, maxprice float64, qtyheld int, tradeon bool) *trailingStopLimit {
	return &trailingStopLimit{
		stopprice:   price - trailingAmt,
		trailingAmt: trailingAmt,
		limitOffset: limitOffset,
		qtyheld:     qtyheld,
		tradeon:     tradeon,
		maxprice:    maxprice,
	}
}

func (tsm *trailingStopLimit) Update(price float64) (*trailingStopLimit, bool) {
	// stop hit, limit sell at stop minus limitOffset
	if price <= tsm.stopprice {
		return tsm, true
	}
	tsm.stopprice = math.Max(tsm.stopprice, price)
	return tsm, false
}
