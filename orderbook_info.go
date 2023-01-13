package orderbookfetcher

import "time"

type OrderbookInfo struct {
	// how many orders does this orderbook contain?
	OrderCount uint
	// how many of those orders are sell orders?
	SellOrderCount uint
	// how many are buy orders
	BuyOrderCount uint
	// name of the location that we fetched the orders for
	LocationName string
	// location id
	LocationID uint64
	// did we fetch those orders from a citadel
	IsCitadel bool
	// when did/does this data expire
	Date time.Time
}

// construct a new instance of the OrderbookInfo
func NewOrderbookInfo(location uint64, expiry time.Time, isCitadel bool) *OrderbookInfo {
	return &OrderbookInfo{
		OrderCount:     0,
		SellOrderCount: 0,
		BuyOrderCount:  0,
		LocationID:     location,
		Date:           expiry,
		IsCitadel:      isCitadel,
	}
}
