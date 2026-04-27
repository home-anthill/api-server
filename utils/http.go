package utils

import (
	"api-server/customerrors"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

const remoteHTTPTimeout = 5 * time.Second

var remoteHTTPClient = &http.Client{
	Timeout: remoteHTTPTimeout,
}

// Get function
func Get(url string) (int, string, error) {
	response, err := remoteHTTPClient.Get(url)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call HTTP GET API of the remote service")
	}
	return readRemoteHTTPResponse(response, "HTTP GET")
}

// Post function
func Post(url string, payloadJSON []byte) (int, string, error) {
	var payloadBody = bytes.NewBuffer(payloadJSON)
	response, err := remoteHTTPClient.Post(url, "application/json", payloadBody)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call HTTP POST API of the remote service")
	}
	return readRemoteHTTPResponse(response, "HTTP POST")
}

// Delete function
func Delete(url string) (int, string, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot create HTTP DELETE request")
	}
	response, err := remoteHTTPClient.Do(req)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call HTTP DELETE API of the remote service")
	}
	return readRemoteHTTPResponse(response, "HTTP DELETE")
}

func readRemoteHTTPResponse(response *http.Response, operation string) (int, string, error) {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, "", customerrors.Wrap(response.StatusCode, err, "Cannot read response body from "+operation)
	}
	bodyString := string(body)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response.StatusCode, bodyString, customerrors.Wrap(
			response.StatusCode,
			fmt.Errorf("%s returned status %d", operation, response.StatusCode),
			"Remote service returned non-success status",
		)
	}
	return response.StatusCode, bodyString, nil
}
