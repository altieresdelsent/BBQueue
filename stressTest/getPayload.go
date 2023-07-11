package stressTest

import (
	"fmt"
	"net/http"
)

const timeoutQueueSeconds = 30

func GetPayload() (int, string) {
	// Make HTTP request
	req, err := http.NewRequest(http.MethodGet, endPoint, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return http.StatusExpectationFailed, ""
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Timeout-Key", fmt.Sprint(timeoutQueueSeconds))

	// Send the request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return http.StatusExpectationFailed, ""
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.Header.Get("Queue-Key")

}