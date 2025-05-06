package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

func New(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("New: error opening file: %w", err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Printf("New: error closing file: %s", err.Error())
		}
	}(file)
	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("New: error decoding file: %w", err)
	}
	if config.FiringLines <= 0 || config.Laps <= 0 || config.LapLen <= 0 || config.PenaltyLen <= 0 {
		return nil, fmt.Errorf("New: invalid config: some fields must be positive")
	}
	return &config, nil
}
