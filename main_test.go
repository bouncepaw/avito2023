package main

import (
	swagger "avito2023/go"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

const host = "http://localhost:8080/"

var client = &http.Client{}

func send[T any](path string, payload any) T {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", host+path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var answer T
	json.NewDecoder(resp.Body).Decode(&answer)
	return answer
}

// I did not call it TestMain because that name has additional connotations.
func TestAPI(t *testing.T) {
	go main()
	time.Sleep(200 * time.Millisecond) // Plenty of time for main() to start.

	send[swagger.InlineResponse200]("create_segment", swagger.CreateSegmentBody{Name: "segment 1"})
	send[swagger.InlineResponse200]("create_segment", swagger.CreateSegmentBody{Name: "segment 2", Percent: 30})
}
