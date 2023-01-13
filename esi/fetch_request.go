package esi

import "time"

type fetchRequest struct {
	// region or citadelid
	LocationID uint64
	// are we fetching citadel or region orders?
	IsCitadel bool
	// when does the endpoint expire?
	Expiry time.Time
	// how often have we skipped fetching the endpoint?
	Skipped int
	// which orderbooks are currently on disk
	FilesWritten []string

	// required by the heap interface
	index int
	// used to delete old orderbooks
	totalWritten uint
}
