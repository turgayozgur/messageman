package waitfor

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// WaitForTruDuration constant.
	ForTruDuration = 5 * time.Second
	// WaitForAPIDuration constant.
	ForAPIDuration = 5 * time.Second
)

// True wait for the given function returns true.
func True(fn func() bool) {
	for {
		success := fn()
		if !success {
			time.Sleep(ForTruDuration)
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
			log.Warn().Err(err).Msg("service is not ready")
			time.Sleep(ForAPIDuration)
			continue
		}
		if response.StatusCode >= 300 {
			log.Warn().Msgf("service is not ready. Non success status code %d on http get. url:%s", response.StatusCode, url)
			time.Sleep(ForAPIDuration)
			continue
		}
		break
	}
}
