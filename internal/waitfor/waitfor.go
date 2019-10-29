package waitfor

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// WaitForTruDuration constant.
	WaitForTruDuration = 5 * time.Second
	// WaitForAPIDuration constant.
	WaitForAPIDuration = 5 * time.Second
)

// True wait for the given function resturns true.
func True(fn func(param []interface{}) bool, params ...interface{}) {
	for {
		success := fn(params)
		if !success {
			time.Sleep(WaitForTruDuration)
			continue
		}
		break
	}
}

// API .
func API(url string) {
	// Clients and Transports are safe for concurrent use by multiple goroutines
	// and for efficiency should only be created once and re-used.
	var httpClient = &http.Client{
		Timeout: time.Second * 60,
	}

	for {
		response, err := httpClient.Get(url)
		if err != nil {
			log.Warn().Err(err).Msg("Main API is not ready")
			time.Sleep(WaitForAPIDuration)
			continue
		}
		if response.StatusCode >= 300 {
			log.Warn().Msgf("Main API is not ready. Non success status code %d on http get. url:%s", response.StatusCode, url)
			time.Sleep(WaitForAPIDuration)
			continue
		}
		break
	}
}
