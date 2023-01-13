package esi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// make a head request to get the expiry for this location
// as we do not care about the data at this time
func (f *Fetcher) GetExpiry(fetchReq *fetchRequest, page uint) (time.Time, error) {
	url := fmt.Sprintf(f.regionURL, fetchReq.LocationID, page)
	if fetchReq.IsCitadel {
		url = fmt.Sprintf(f.citadelURL, fetchReq.LocationID, page)
	}

	headReq, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return time.Time{}, err
	}

	if fetchReq.IsCitadel {
		headReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.accessToken))
	}

	resp, err := f.client.Do(headReq)
	if err != nil {
		return time.Time{}, err
	} else if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("request returned status code %s", resp.Status)
	}

	return time.Parse(time.RFC1123, resp.Header.Get("Expires"))
}

// fetch every order page for the given location
// uses a callback function to save memory + allocations
func (f *Fetcher) GetOrders(fetchReq *fetchRequest, cb func(*fetchResponse, uint)) error {
	fetchResp := &fetchResponse{
		LocationID: fetchReq.LocationID,
		IsCitadel:  fetchReq.IsCitadel,
	}
	// what page are we currently fetching?
	page := 1
	// how often have we tried this page?
	retries := 0
	for {
		// construct the url
		url := fmt.Sprintf(f.regionURL, fetchReq.LocationID, page)
		if fetchReq.IsCitadel {
			url = fmt.Sprintf(f.citadelURL, fetchReq.LocationID, page)
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		if fetchReq.IsCitadel {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.accessToken))
		}

		resp, err := f.client.Do(req)
		if err != nil {
			return err
		} else if resp.StatusCode != http.StatusOK {
			// retry the request if it failed
			if retries < 3 {
				retries++
				log.Printf("request for page %d returned status code %s; retrying...", page, resp.Status)
				continue
			} else {
				return fmt.Errorf("request for page %d returned status code %s", page, resp.Status)
			}
		}
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&fetchResp.Orders); err != nil {
			return err
		}

		fetchResp.Expiry, err = time.Parse(time.RFC1123, resp.Header.Get("Expires"))
		if err != nil {
			return err
		}
		cb(fetchResp, uint(page))

		if len(fetchResp.Orders) < 1000 {
			break
		}
		retries = 0
		page++
	}
	return nil
}

// retrieve the name of a location
func (f *Fetcher) GetLocationName(location uint64, isCitadel bool) (string, error) {
	// are we pulling the name of a citadel or a region?
	url := fmt.Sprintf("https://esi.evetech.net/latest/universe/regions/%d/?datasource=tranquility", location)
	if isCitadel {
		url = fmt.Sprintf("https://esi.evetech.net/latest/universe/structures/%d/?datasource=tranquility", location)
	}
	log.Print(url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	if isCitadel {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.accessToken))
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request returned status code: %s", resp.Status)
	}
	defer resp.Body.Close()

	// anonymous struct to extract just the name from the json response
	locationName := struct {
		Name string `json:"name"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&locationName); err != nil {
		return "", err
	}
	return locationName.Name, nil
}
