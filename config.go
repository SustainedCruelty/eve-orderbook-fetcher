package orderbookfetcher

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	// how many orderbooks are we saving on disk (per location)
	RetentionPeriod uint `json:"retentionPeriod"`
	// how often are we fetching?
	Interval uint `json:"interval"`
	// what regions are we fetching?
	Regions []uint64 `json:"regions"`
	// what citadels are we fetching?
	Citadels []uint64 `json:"citadels"`
	// client id for the esi application
	ClientID string `json:"clientId"`
	// refresh token to retrieve our access token
	RefreshToken string `json:"refreshToken"`
}

// load a configuration from a text file
func LoadConfiguration(fileName string) (*Configuration, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var config *Configuration
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
