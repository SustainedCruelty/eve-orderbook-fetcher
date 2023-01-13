package esi

import (
	"container/heap"
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	orderbookfetcher "github.com/SustainedCruelty/eve-orderbook-fetcher"
)

type Fetcher struct {
	// http client used to make the esi requests
	client *http.Client

	// cancellation function to stop program execution
	cancel context.CancelFunc
	// wait for goroutines to finish
	wg sync.WaitGroup

	// base URLs for fetching region or citadel orders
	citadelURL string
	regionURL  string

	// holds our requests
	pq PriorityQueue

	// configuration
	config *orderbookfetcher.Configuration

	// ESI access token used to make authenticated requests
	accessToken string
	tokenExpiry time.Time

	// look up the name of a structure or region by id
	Locations map[uint64]string
	// look up information about the orderbook by filename
	WrittenOrderbooks map[string]*orderbookfetcher.OrderbookInfo
}

func NewFetcher(config *orderbookfetcher.Configuration) *Fetcher {
	return &Fetcher{
		config:            config,
		Locations:         make(map[uint64]string, len(config.Regions)+len(config.Citadels)),
		WrittenOrderbooks: make(map[string]*orderbookfetcher.OrderbookInfo),
		client:            http.DefaultClient,

		citadelURL: "https://esi.evetech.net/latest/markets/structures/%d/?datasource=tranquility&page=%d",
		regionURL:  "https://esi.evetech.net/latest/markets/%d/orders/?datasource=tranquility&order_type=all&page=%d",
	}
}

func (f *Fetcher) Start() (err error) {
	// context required for cancellation
	var ctx context.Context
	ctx, f.cancel = context.WithCancel(context.Background())

	// queue holds as many elements as we have locations
	queueLength := len(f.config.Regions) + len(f.config.Citadels)
	f.pq = make(PriorityQueue, queueLength)

	// get our initial access token
	// before we start the goroutine
	if len(f.config.Citadels) > 0 {
		tokens, err := f.RefreshToken()
		if err != nil {
			log.Printf("failed to refresh tokens: %s", err)
			return err
		}
		f.accessToken = tokens.AccessToken
		f.tokenExpiry = time.Now().Add(time.Second * time.Duration(tokens.ExpiresIn-5))
	}

	// create requests for the locations to be fetched
	// and add them to a priority queue
	for i, location := range append(f.config.Citadels, f.config.Regions...) {

		// are we in the first or second slice?
		isCitadel := i < len(f.config.Citadels)

		// fetch the location name for every location
		// and add them to the map
		_, ok := f.Locations[location]
		if !ok {
			locName, err := f.GetLocationName(location, isCitadel)
			if err != nil {
				return err
			}
			log.Printf("%d - %s", location, locName)
			f.Locations[location] = locName
		}

		// construct the request
		// and put it in the queue
		f.pq[i] = &fetchRequest{
			LocationID:   location,
			IsCitadel:    isCitadel,
			Expiry:       time.Now(),
			Skipped:      -1,
			FilesWritten: make([]string, f.config.RetentionPeriod),
			totalWritten: 0,
		}
	}

	// sort the priority queue
	heap.Init(&f.pq)

	// work on the requests
	f.wg.Add(1)
	go func() {
		f.worker(ctx)
		f.wg.Done()
	}()

	// only keep the token refreshed
	// if we have to request citadel orders
	if len(f.config.Citadels) > 0 {
		f.wg.Add(1)
		go func() {
			f.tokenRefresher(ctx)
			f.wg.Done()
		}()
	}

	return nil
}

func (f *Fetcher) Shutdown() {
	log.Println("shutting down...")
	// cancel the context and wait for execution to finish
	// so we aren't left with open/unfinished files
	f.cancel()
	f.wg.Wait()
	log.Println("done!")
}

func (f *Fetcher) worker(ctx context.Context) {
	for len(f.pq) > 0 {
		// get our request with the earliest expiry from the heap
		request := heap.Pop(&f.pq).(*fetchRequest)
		// wait until it expires
		wait := time.Until(request.Expiry) + time.Second*1
		log.Printf("location %d expires in %s", request.LocationID, wait)
		select {
		case <-time.After(wait):
			// are we skipping or fetching?
			if request.Skipped != -1 && f.config.Interval > uint(request.Skipped+1) {
				var err error
				request.Expiry, err = f.GetExpiry(request, 1)
				if err != nil {
					log.Printf("failed to fetch the expiry: %s", err)
					return
				}
				request.Skipped++
				heap.Push(&f.pq, request)
				continue

			} else {
				log.Printf("fetching location %d", request.LocationID)
				// csv file containing the orderbook
				var file *os.File
				// some stats about the orderbook
				var info *orderbookfetcher.OrderbookInfo
				if err := f.GetOrders(request, func(fr *fetchResponse, page uint) {
					// are we making a new orderbook or writing to an existing one?
					if page == 1 {
						var err error
						file, info, err = fr.CreateNewCSV()
						if err != nil {
							log.Printf("failed to create file: %s", err)
						}
						request.Expiry = fr.Expiry
						info.LocationName = f.Locations[request.LocationID]
					} else {
						fr.WriteToExistingCSV(file, info)
					}

				}); err != nil {
					log.Printf("failed to fetch orders: %s", err)
				}

				// close the file
				if err := file.Close(); err != nil {
					log.Printf("failed to close: %s", err)
					return
				}

				// remove the tmp file as we are done writing
				fileName := strings.Trim(file.Name(), ".tmp")
				if err := os.Rename(file.Name(), fileName); err != nil {
					log.Printf("failed to rename: %s", err)
					return
				}

				// do we have to delete an old orderbook?
				if request.totalWritten < f.config.RetentionPeriod {
					// add the file to the slice
					request.FilesWritten[request.totalWritten] = fileName
					request.totalWritten++
				} else if f.config.RetentionPeriod > 0 {
					// index of the file to be removed
					idx := request.totalWritten % f.config.RetentionPeriod
					// remove the file from disk
					os.Remove(request.FilesWritten[idx])
					log.Printf("removed file: %s", request.FilesWritten[idx])
					// remove it from the map
					delete(f.WrittenOrderbooks, request.FilesWritten[idx])
					// overwrite it with the new file
					request.FilesWritten[idx] = fileName
					request.totalWritten++
				}
				// put the info into the map
				f.WrittenOrderbooks[fileName] = info
				log.Printf("finished fetching location %d", request.LocationID)
				request.Skipped = 0

				// add the request back to the heap
				heap.Push(&f.pq, request)
			}

		case <-ctx.Done():
			return
		}
	}
}

// refreshes the access token every 20 minutes
func (f *Fetcher) tokenRefresher(ctx context.Context) {
	for {
		select {
		// wait for the token to expiry
		case <-time.After(time.Until(f.tokenExpiry)):
			tokens, err := f.RefreshToken()
			if err != nil {
				log.Printf("failed to fetch tokens: %s", err)
				return
			}
			// set the new token + expiry
			f.accessToken = tokens.AccessToken
			f.tokenExpiry = time.Now().Add(time.Second * time.Duration(tokens.ExpiresIn-5))

			log.Println("refreshed token")

		case <-ctx.Done():
			return
		}
	}
}
