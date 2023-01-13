package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	orderbookfetcher "github.com/SustainedCruelty/eve-orderbook-fetcher"
	"github.com/SustainedCruelty/eve-orderbook-fetcher/esi"
	"github.com/SustainedCruelty/eve-orderbook-fetcher/http"
)

func main() {
	// used to terminate execution
	ctx, cancel := context.WithCancel(context.Background())
	// cancel the context if the program is terminated
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	// load our config file
	config, err := orderbookfetcher.LoadConfiguration("config.json")
	if err != nil {
		log.Fatalf("failed to load the configuration: %s", err)
	}

	log.Printf("fetching %d citadel(s) and %d regions(s)", len(config.Citadels), len(config.Regions))

	m := NewMain(config)

	// start the server + fetcher
	if err := m.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// wait for the program to terminate
	<-ctx.Done()

	// graceful shutdown
	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Main struct {
	// config file
	Configuration *orderbookfetcher.Configuration
	// fetches the orderbooks
	Fetcher *esi.Fetcher
	// serves a small ui
	Server *http.Server
}

// construct a new main object that holds our instances
func NewMain(config *orderbookfetcher.Configuration) *Main {
	return &Main{
		Configuration: config,
		Fetcher:       esi.NewFetcher(config),
		Server:        http.NewServer(),
	}
}

// run our services and inject the dependencies
func (m *Main) Run(ctx context.Context) error {
	log.Println("running...")
	if err := m.Fetcher.Start(); err != nil {
		return err
	}
	m.Server.ESIFetcher = m.Fetcher
	if err := m.Server.Open(); err != nil {
		return err
	}
	return nil
}

// graceful shutdown
func (m *Main) Close() error {
	m.Fetcher.Shutdown()
	if err := m.Server.Close(); err != nil {
		return err
	}
	return nil
}
