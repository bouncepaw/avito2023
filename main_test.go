package main

import (
	swagger "avito2023/go"
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

const host = "http://localhost:8080/"

var client = &http.Client{}

func post[T any](path string, payload any) T {
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
	response := post[T](path, payload)
	return response, reflect.DeepEqual(response, expect)
}

func TestCreateSegment(t *testing.T) {
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

func TestDeleteSegment(t *testing.T) {
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

func TestUpdateUser(t *testing.T) {
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

func TestGetSegments(t *testing.T) {
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

func TestHistoryPost(t *testing.T) {
	table := []struct {
		payload          swagger.HistoryBody
		expectedResponse swagger.InlineResponse2002
	}{
		{
			swagger.HistoryBody{1000, 1},
			swagger.InlineResponse2002{Status: "error", Error_: "bad time"},
		},
		{
			swagger.HistoryBody{2024, 13},
			swagger.InlineResponse2002{Status: "error", Error_: "bad time"},
		},
		{
			swagger.HistoryBody{2023, 6},
			swagger.InlineResponse2002{Status: "ok", Link: "/history?year=2023&month=6"},
		},
		{
			swagger.HistoryBody{2023, 10},
			swagger.InlineResponse2002{Status: "ok", Link: "/history?year=2023&month=10"},
		},
	}
	for i, test := range table {
		response, ok := yesbut("history", test.payload, test.expectedResponse)
		if !ok {
			t.Errorf("Failed test %d: got %q instead of %q", i, response, test.expectedResponse)
		}
	}
}

func TestHistoryGet(t *testing.T) {

}

// I did not call it TestMain because that name has additional connotations.
func TestMain(m *testing.M) {
	go main()
	time.Sleep(200 * time.Millisecond) // Plenty of time for main() to start.
	os.Exit(m.Run())
}

func todayYearMonth() (year int, month int) {
	t := time.Now()
	return t.Year(), int(t.Month())
}
