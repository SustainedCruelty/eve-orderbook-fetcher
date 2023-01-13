package http

import (
	"context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/SustainedCruelty/eve-orderbook-fetcher/esi"
)

type Server struct {
	ln     net.Listener
	server *http.Server
	router *http.ServeMux

	ESIFetcher *esi.Fetcher
}

// Create a new instance of our server
func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: http.DefaultServeMux,
	}

	// register all of the necessary handlers
	s.registerOrderbookRoutes(s.router)
	s.router.Handle("/orderbooks/", http.StripPrefix("/orderbooks", http.FileServer(http.Dir("./orderbooks"))))
	s.router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "http/assets/favicon.ico")
	})

	return s
}

// start listening
func (s *Server) Open() (err error) {
	if s.ln, err = net.Listen("tcp", ":8080"); err != nil {
		return err
	}
	log.Println("listening on port 8080")
	go s.server.Serve(s.ln)
	return nil
}

// close the server
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	return s.server.Shutdown(ctx)
}
