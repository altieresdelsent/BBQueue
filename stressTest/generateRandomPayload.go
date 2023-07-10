package stressTest

import "math/rand"

type Payload struct {
	StringFields []string `json:"string_fields"`
	IntFields    []int    `json:"int_fields"`
	BoolFields   []bool   `json:"bool_fields"`
}

func generateRandomPayload() Payload {
	payload := Payload{}

	// Generate 12 random strings
	for i := 0; i < 12; i++ {
		payload.StringFields = append(payload.StringFields, generateRandomString())
	}

	// Generate 4 random integers
	for i := 0; i < 4; i++ {
		payload.IntFields = append(payload.IntFields, generateRandomInt())
	}

	// Generate 2 random booleans
	for i := 0; i < 2; i++ {
		payload.BoolFields = append(payload.BoolFields, generateRandomBool())
	}

	return payload
}

func generateRandomString() string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 10

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

func generateRandomInt() int {
	return rand.Intn(100)
}

func generateRandomBool() bool {
	return rand.Intn(2) == 1
}
