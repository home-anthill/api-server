package utils

import (
	"api-server/customerrors"
	"bytes"
	"io"
	"net/http"
)

// Get function
func Get(url string) (int, string, error) {
	response, err := http.Get(url)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call HTTP GET API of the remote service")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, "", customerrors.Wrap(response.StatusCode, err, "Cannot read response body from HTTP GET")
	}
	return response.StatusCode, string(body), nil
}

// Post function
func Post(url string, payloadJSON []byte) (int, string, error) {
	var payloadBody = bytes.NewBuffer(payloadJSON)
	response, err := http.Post(url, "application/json", payloadBody)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call HTTP POST API of the remote service")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, "", customerrors.Wrap(response.StatusCode, err, "Cannot read response body from HTTP POST")
	}
	return response.StatusCode, string(body), nil
}
