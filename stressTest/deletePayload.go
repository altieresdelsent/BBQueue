package stressTest

import (
	"fmt"
	"net/http"
	"time"
)

const endPointProcessing = "http://127.0.0.1:8080/processing"

func DeletePayload(key string) int {

	// Make HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprint(endPointProcessing, "/", key), nil)
	if err != nil {
		fmt.Println("Error DeletePayload NewRequest:", err)
		return http.StatusExpectationFailed
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := http.Client{
		Timeout: time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error DeletePayload client.Do:", err)
		return http.StatusExpectationFailed
	}
	defer resp.Body.Close()
	return resp.StatusCode

}
