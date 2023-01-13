package http

import (
	"html/template"
	"log"
	"net/http"

	orderbookfetcher "github.com/SustainedCruelty/eve-orderbook-fetcher"
)

func (s *Server) registerOrderbookRoutes(r *http.ServeMux) {
	r.HandleFunc("/", s.handleOrderbookInfo)
}

// serve the template
func (s *Server) handleOrderbookInfo(w http.ResponseWriter, r *http.Request) {
	// load the file
	tmpl, err := template.ParseFiles("http/assets/index.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("failed to parse the template: %s", err)
		return
	}

	// throwaway struct containing the data to be displayed
	data := struct {
		Locations  map[uint64]string
		Orderbooks map[string]*orderbookfetcher.OrderbookInfo
	}{
		Locations:  s.ESIFetcher.Locations,
		Orderbooks: s.ESIFetcher.WrittenOrderbooks,
	}

	// serve the template
	if err = tmpl.Execute(w, data); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Fatalf("failed to execute the template: %s", err)
	}
}
