package stressTest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const endPointQueue = "http://127.0.0.1:8080/queue"

func PostRandomPayload() int {
	payload := generateRandomPayload()
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error:", err)
		return http.StatusExpectationFailed
	}

	// Make HTTP request
	req, err := http.NewRequest(http.MethodPost, endPointQueue, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error:", err)
		return http.StatusExpectationFailed
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return http.StatusExpectationFailed
	}
	defer resp.Body.Close()
	return resp.StatusCode

}
