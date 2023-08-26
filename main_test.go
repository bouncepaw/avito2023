package main

import (
	swagger "avito2023/go"
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
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

func yesbut[T any](path string, payload any, expect T) (T, bool) {
	response := send[T](path, payload)
	return response, reflect.DeepEqual(response, expect)
}

func testCreateSegment(t *testing.T) {
	table := []struct {
		payload          swagger.CreateSegmentBody
		expectedResponse swagger.InlineResponse200
	}{
		{
			swagger.CreateSegmentBody{Name: "segment 1"},
			swagger.InlineResponse200{Status: "ok"},
		},
		{
			swagger.CreateSegmentBody{Name: "segment 2", Percent: 30},
			swagger.InlineResponse200{Status: "ok"},
		},
		{
			swagger.CreateSegmentBody{Name: "segment 1"},
			swagger.InlineResponse200{Status: "error", Error_: "name taken"},
		},
		{
			swagger.CreateSegmentBody{Name: "segment to delete 1"},
			swagger.InlineResponse200{Status: "ok"},
		},
		{
			swagger.CreateSegmentBody{Name: "quarnishone", Percent: 123},
			swagger.InlineResponse200{Status: "error", Error_: "bad percent"},
		},
	}
	for i, test := range table {
		response, ok := yesbut("create_segment", test.payload, test.expectedResponse)
		if !ok {
			t.Errorf("Failed test %d: got %q instead of %q", i, response, test.expectedResponse)
		}
	}
}

func testDeleteSegment(t *testing.T) {
	table := []struct {
		payload          swagger.DeleteSegmentBody
		expectedResponse swagger.InlineResponse200
	}{
		{
			swagger.DeleteSegmentBody{Name: "segment to delete 1"},
			swagger.InlineResponse200{Status: "ok"},
		},
		{
			swagger.DeleteSegmentBody{Name: "segment to delete 1"},
			swagger.InlineResponse200{Status: "error", Error_: "already deleted"},
		},
		{
			swagger.DeleteSegmentBody{Name: "quasimodo"},
			swagger.InlineResponse200{Status: "error", Error_: "name free"},
		},
	}
	for i, test := range table {
		response, ok := yesbut("delete_segment", test.payload, test.expectedResponse)
		if !ok {
			t.Errorf("Failed test %d: got %q instead of %q", i, response, test.expectedResponse)
		}
	}
}

func testUpdateUser(t *testing.T) {
	table := []struct {
		payload          swagger.UpdateUserBody
		expectedResponse swagger.InlineResponse200
	}{
		{
			swagger.UpdateUserBody{Id: 101, AddToSegments: []string{"segment 1", "segment 2"}, RemoveFromSegments: []string{}},
			swagger.InlineResponse200{Status: "ok"},
		},
		{
			swagger.UpdateUserBody{Id: 101, AddToSegments: []string{}, RemoveFromSegments: []string{"segment 2"}},
			swagger.InlineResponse200{Status: "ok"},
		},
	}
	for i, test := range table {
		response, ok := yesbut("update_user", test.payload, test.expectedResponse)
		if !ok {
			t.Errorf("Failed test %d: got %q instead of %q", i, response, test.expectedResponse)
		}
	}
}

func testGetSegments(t *testing.T) {
	table := []struct {
		payload          swagger.GetSegmentsBody
		expectedResponse swagger.InlineResponse2001
	}{
		{
			swagger.GetSegmentsBody{Id: 101},
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"segment 1"}},
		},
		{
			swagger.GetSegmentsBody{Id: 10},
			swagger.InlineResponse2001{Status: "ok"},
		},
	}
	for i, test := range table {
		response, ok := yesbut("get_segments", test.payload, test.expectedResponse)
		if !ok {
			t.Errorf("Failed test %d: got %q instead of %q", i, response, test.expectedResponse)
		}
	}
}

// I did not call it TestMain because that name has additional connotations.
func TestAPI(t *testing.T) {
	go main()
	time.Sleep(200 * time.Millisecond) // Plenty of time for main() to start.
	testCreateSegment(t)
	testDeleteSegment(t)
	testUpdateUser(t)
	testGetSegments(t)
}
