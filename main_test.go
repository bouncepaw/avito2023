package main

import (
	swagger "avito2023/go"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type Testable interface {
	Test(idx int, t *testing.T)
}

type TestCreate struct {
	swagger.CreateSegmentBody
	swagger.InlineResponse200
}

func (tc *TestCreate) Test(idx int, t *testing.T) {
	response, ok := yesbut("create_segment", tc.CreateSegmentBody, tc.InlineResponse200)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.InlineResponse200)
	}
}

type TestDelete struct {
	swagger.DeleteSegmentBody
	swagger.InlineResponse200
}

func (tc *TestDelete) Test(idx int, t *testing.T) {
	response, ok := yesbut("delete_segment", tc.DeleteSegmentBody, tc.InlineResponse200)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.InlineResponse200)
	}
}

type TestUpdate struct {
	swagger.UpdateUserBody
	swagger.InlineResponse200
}

func (tc *TestUpdate) Test(idx int, t *testing.T) {
	response, ok := yesbut("update_user", tc.UpdateUserBody, tc.InlineResponse200)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.InlineResponse200)
	}
}

type TestGet struct {
	swagger.GetSegmentsBody
	possibility1 swagger.InlineResponse2001
	possibility2 swagger.InlineResponse2001
}

func (tc *TestGet) Test(idx int, t *testing.T) {
	var r any = post[swagger.InlineResponse2001]("get_segments", tc.GetSegmentsBody)
	response := r.(swagger.InlineResponse2001)

	ok1 := reflect.DeepEqual(response, tc.possibility1)
	ok2 := reflect.DeepEqual(response, tc.possibility2)
	if !(ok1 || ok2) {
		t.Errorf("Failed test %d: got %q instead of %q or %q", idx, response, tc.possibility1, tc.possibility2)
	}
}

type TestHistory struct {
	swagger.HistoryBody
	swagger.InlineResponse2002
}

func (tc *TestHistory) Test(idx int, t *testing.T) {
	response, ok := yesbut("history", tc.HistoryBody, tc.InlineResponse2002)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.InlineResponse2002)
	}
}

type TestWait int

func (tc TestWait) Test(idx int, t *testing.T) {
	<-time.After(time.Duration(tc) * time.Millisecond)
	log.Println("Waited", tc, "milliseconds")
}

func TestCreateSegment(t *testing.T) {
	for i, test := range []Testable{
		&TestCreate{
			swagger.CreateSegmentBody{Name: "segment 1"},
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "segment 2"},
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "segment 1"},
			swagger.InlineResponse200{Status: "error", Error_: "name taken"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "segment to delete 1"},
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "quarnishone", Percent: 123},
			swagger.InlineResponse200{Status: "error", Error_: "bad percent"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "fifty-fifty", Percent: 50}, // !
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "hundred", Percent: 100}, // !
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: ""},
			swagger.InlineResponse200{Status: "error", Error_: "name empty"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestDeleteSegment(t *testing.T) {
	for i, test := range []Testable{
		&TestDelete{
			swagger.DeleteSegmentBody{Name: "segment to delete 1"},
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestDelete{
			swagger.DeleteSegmentBody{Name: "segment to delete 1"},
			swagger.InlineResponse200{Status: "error", Error_: "already deleted"},
		},
		&TestDelete{
			swagger.DeleteSegmentBody{Name: "quasimodo"},
			swagger.InlineResponse200{Status: "error", Error_: "name free"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestUpdateUser(t *testing.T) {
	for i, test := range []Testable{
		&TestUpdate{
			swagger.UpdateUserBody{Id: 101, AddToSegments: []string{"segment 1", "segment 2"}, RemoveFromSegments: []string{}},
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestUpdate{
			swagger.UpdateUserBody{Id: 101, AddToSegments: []string{}, RemoveFromSegments: []string{"segment 2"}},
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestUpdate{
			// This shall perish in a second
			swagger.UpdateUserBody{Id: 1234, AddToSegments: []string{"segment 1"}, Ttl: 1},
			swagger.InlineResponse200{Status: "ok"},
		},
		&TestGet{
			swagger.GetSegmentsBody{Id: 1234},
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"segment 1", "hundred"}}, // The segment has not yet perished
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"segment 1", "fifty-fifty", "hundred"}},
		},
		&TestUpdate{
			// This shall perish in a second
			swagger.UpdateUserBody{Id: 12345, AddToSegments: []string{"segment 1", "segment 2"}, Ttl: 1},
			swagger.InlineResponse200{Status: "ok"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestGetSegments(t *testing.T) {
	for i, test := range []Testable{
		TestWait(1100), // + 100 ms just in case
		&TestGet{
			swagger.GetSegmentsBody{Id: 101},
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"segment 1", "hundred"}},
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"segment 1", "fifty-fifty", "hundred"}},
		},
		&TestGet{
			swagger.GetSegmentsBody{Id: 10},
			swagger.InlineResponse2001{Status: "ok"}, // We didn't know of this one before, so not in any segment,
			swagger.InlineResponse2001{Status: "ok"}, // not even in "hundred".
		},
		&TestGet{
			swagger.GetSegmentsBody{Id: 1234},
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"hundred"}}, // The segment perished after a second
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"fifty-fifty", "hundred"}},
		},
		&TestGet{
			swagger.GetSegmentsBody{Id: 12345},
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"hundred"}}, // The segment perished after a second
			swagger.InlineResponse2001{Status: "ok", Segments: []string{"fifty-fifty", "hundred"}},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestHistoryPost(t *testing.T) {
	for i, test := range []Testable{
		&TestHistory{
			swagger.HistoryBody{1000, 1},
			swagger.InlineResponse2002{Status: "error", Error_: "bad time"},
		},
		&TestHistory{
			swagger.HistoryBody{2024, 13},
			swagger.InlineResponse2002{Status: "error", Error_: "bad time"},
		},
		&TestHistory{
			swagger.HistoryBody{2023, 6},
			swagger.InlineResponse2002{Status: "ok", Link: "/history?year=2023&month=6"},
		},
		&TestHistory{
			swagger.HistoryBody{2023, 10},
			swagger.InlineResponse2002{Status: "ok", Link: "/history?year=2023&month=10"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestHistoryGet(t *testing.T) {
	times := time.Now()
	todayYear, todayMonth := times.Year(), int(times.Month())

	table := []struct {
		// We compare line count instead of string equality because timestamps are too much hassle.
		year, month, status int
		wantManyLines       bool
	}{
		// Found empirically. Increment for every new operation in other tests.
		{todayYear, todayMonth, 200, false},
		{2023, 4, 200, true},
		{-15, 14, 404, true},
	}
	for i, test := range table {
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf(`%shistory?year=%d&month=%d`, host, test.year, test.month),
			http.NoBody)
		if err != nil {
			panic(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		// Take random additions into account
		if linecnt := bytes.Count(b, []byte{'\n'}); linecnt > 1 && test.wantManyLines {
			t.Errorf("Failed test %d: got %d, wanted many = %v. CSV: %s", i, linecnt, test.wantManyLines, string(b))
		}
	}
}

// I did not call it TestMain because that name has additional connotations.
func TestMain(m *testing.M) {
	go main()
	time.Sleep(200 * time.Millisecond) // Plenty of time for main() to start.
	os.Exit(m.Run())
}
