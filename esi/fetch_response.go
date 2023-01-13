package esi

import (
	"fmt"
	"os"
	"time"

	orderbookfetcher "github.com/SustainedCruelty/eve-orderbook-fetcher"
)

type fetchResponse struct {
	// the actual orders
	Orders []*orderbookfetcher.MarketOrder
	// when does this data expire?
	Expiry time.Time
	// what location is this data for?
	LocationID uint64
	// are those citadel or regional orders?
	IsCitadel bool
}

// create a new orderbook csv file and write the current order page to it
func (r *fetchResponse) CreateNewCSV() (*os.File, *orderbookfetcher.OrderbookInfo, error) {
	// create the csv file, made up of location and expiry timestamp
	fileName := fmt.Sprintf("orderbooks/%d_%d.csv", r.LocationID, r.Expiry.Unix())
	file, err := os.Create(fileName + ".tmp")
	if err != nil {
		return nil, nil, err
	}
	// create some stats about the orders we are writing
	info := orderbookfetcher.NewOrderbookInfo(r.LocationID, r.Expiry, r.IsCitadel)
	// write the column names
	_, err = fmt.Fprintln(file, "ORDERID,TYPEID,SYSTEMID,LOCATIONID,PRICE,RANGE,ISBUY,ISSUED,DURATION,MINVOLUME,VOLUMEREMAIN,VOLUMETOTAL")
	if err != nil {
		return nil, nil, err
	}

	// write all of the orders
	for _, order := range r.Orders {
		info.OrderCount++
		if order.IsBuyOrder {
			info.BuyOrderCount++
		} else {
			info.SellOrderCount++
		}
		order.WriteAsCSV(file)
	}
	return file, info, nil
}

// write a page of orders to an already exisitng orderbook file
func (r *fetchResponse) WriteToExistingCSV(file *os.File, info *orderbookfetcher.OrderbookInfo) {
	for _, order := range r.Orders {
		info.OrderCount++
		if order.IsBuyOrder {
			info.BuyOrderCount++
		} else {
			info.SellOrderCount++
		}
		order.WriteAsCSV(file)
	}
}
