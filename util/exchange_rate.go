package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type ExhangeResponse struct {
	Result string             `json:"result"`
	Date   string             `json:"time_last_update_utc"`
	Rate   map[string]float64 `json:"conversion_rates"`
}

func GetExchangeAPI() (*ExhangeResponse, error) {
	client := http.Client{Timeout: 5 * time.Second}

	config, err := LoadConfig("..")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
		return nil, err
	}
	resp, err := client.Get(config.ExchangeAPI)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot get exchange api")
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("get url failed: %s", resp.Status)
	}

	var result ExhangeResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decode json failed :%s", err)
	}

	return &result, nil

}
