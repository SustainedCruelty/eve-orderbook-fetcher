package orderbookfetcher

import (
	"fmt"
	"io"
	"time"
)

type MarketOrder struct {
	Duration     int32     `json:"duration"`
	IsBuyOrder   bool      `json:"is_buy_order"`
	Issued       time.Time `json:"issued"`
	LocationID   int64     `json:"location_id"`
	MinVolume    int32     `json:"min_volume"`
	OrderID      int64     `json:"order_id"`
	Price        float32   `json:"price"`
	Range        string    `json:"range"`
	SystemID     int32     `json:"system_id"`
	TypeID       int32     `json:"type_id"`
	VolumeRemain int32     `json:"volume_remain"`
	VolumeTotal  int32     `json:"volume_total"`
}

// write the MarketOrder as a new line to an existing csv file
func (order *MarketOrder) WriteAsCSV(writer io.Writer) {
	row := fmt.Sprintf("%d,%d,%d,%d,%f,%s,%v,%d,%d,%d,%d,%d",
		order.OrderID,
		order.TypeID,
		order.SystemID,
		order.LocationID,
		order.Price,
		order.Range,
		order.IsBuyOrder,
		order.Issued.UTC().Unix(),
		order.Duration,
		order.MinVolume,
		order.VolumeRemain,
		order.VolumeTotal,
	)
	fmt.Fprintln(writer, row)
}
