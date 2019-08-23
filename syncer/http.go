package sync

import (
	"net/http"
)

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: httpTimeout,
	}
}
