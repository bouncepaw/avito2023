package main

import (
	"avito2023/web"
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
	web.CreateSegmentBody
	web.ResponseUsual
}

func (tc *TestCreate) Test(idx int, t *testing.T) {
	response, ok := yesbut("create_segment", tc.CreateSegmentBody, tc.ResponseUsual)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.ResponseUsual)
	}
}

type TestDelete struct {
	web.DeleteSegmentBody
	web.ResponseUsual
}

func (tc *TestDelete) Test(idx int, t *testing.T) {
	response, ok := yesbut("delete_segment", tc.DeleteSegmentBody, tc.ResponseUsual)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.ResponseUsual)
	}
}

type TestUpdate struct {
	web.UpdateUserBody
	web.ResponseUsual
}

func (tc *TestUpdate) Test(idx int, t *testing.T) {
	response, ok := yesbut("update_user", tc.UpdateUserBody, tc.ResponseUsual)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.ResponseUsual)
	}
}

type TestGet struct {
	web.GetSegmentsBody
	possibility1 web.ResponseGetSegments
	possibility2 web.ResponseGetSegments
}

func (tc *TestGet) Test(idx int, t *testing.T) {
	response := post[web.ResponseGetSegments]("get_segments", tc.GetSegmentsBody)

	ok1 := reflect.DeepEqual(response, tc.possibility1)
	ok2 := reflect.DeepEqual(response, tc.possibility2)
	if !(ok1 || ok2) {
		t.Errorf("Failed test %d: got %q instead of %q or %q", idx, response, tc.possibility1, tc.possibility2)
	}
}

type TestHistory struct {
	web.HistoryBody
	web.ResponseHistory
}

func (tc *TestHistory) Test(idx int, t *testing.T) {
	response, ok := yesbut("history", tc.HistoryBody, tc.ResponseHistory)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.ResponseHistory)
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
			web.CreateSegmentBody{Name: "segment 1"},
			web.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			web.CreateSegmentBody{Name: "segment 2"},
			web.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			web.CreateSegmentBody{Name: "segment 1"},
			web.ResponseUsual{Status: "error", Err: "name taken"},
		},
		&TestCreate{
			web.CreateSegmentBody{Name: "segment to delete 1"},
			web.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			web.CreateSegmentBody{Name: "quarnishone", Percent: 123},
			web.ResponseUsual{Status: "error", Err: "bad percent"},
		},
		&TestCreate{
			web.CreateSegmentBody{Name: "fifty-fifty", Percent: 50}, // !
			web.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			web.CreateSegmentBody{Name: "hundred", Percent: 100}, // !
			web.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			web.CreateSegmentBody{Name: ""},
			web.ResponseUsual{Status: "error", Err: "name empty"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestDeleteSegment(t *testing.T) {
	for i, test := range []Testable{
		&TestDelete{
			web.DeleteSegmentBody{Name: "segment to delete 1"},
			web.ResponseUsual{Status: "ok"},
		},
		&TestDelete{
			web.DeleteSegmentBody{Name: "segment to delete 1"},
			web.ResponseUsual{Status: "error", Err: "segment deleted"},
		},
		&TestDelete{
			web.DeleteSegmentBody{Name: "quasimodo"},
			web.ResponseUsual{Status: "error", Err: "name free"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestUpdateUser(t *testing.T) {
	for i, test := range []Testable{
		&TestUpdate{
			web.UpdateUserBody{Id: 546, AddToSegments: []string{"segment to delete 1"}},
			web.ResponseUsual{Status: "error", Err: "segment deleted"},
		},
		&TestUpdate{
			web.UpdateUserBody{Id: 101, AddToSegments: []string{"segment 1", "segment 2"}, RemoveFromSegments: []string{}},
			web.ResponseUsual{Status: "ok"},
		},
		&TestUpdate{
			web.UpdateUserBody{Id: 101, AddToSegments: []string{}, RemoveFromSegments: []string{"segment 2"}},
			web.ResponseUsual{Status: "ok"},
		},
		&TestUpdate{
			// This shall perish in a second
			web.UpdateUserBody{Id: 1234, AddToSegments: []string{"segment 1"}, Ttl: 1},
			web.ResponseUsual{Status: "ok"},
		},
		&TestGet{
			web.GetSegmentsBody{Id: 1234},
			web.ResponseGetSegments{Status: "ok", Segments: []string{"segment 1", "hundred"}}, // The segment has not yet perished
			web.ResponseGetSegments{Status: "ok", Segments: []string{"segment 1", "fifty-fifty", "hundred"}},
		},
		&TestUpdate{
			// This shall perish in a second
			web.UpdateUserBody{Id: 12345, AddToSegments: []string{"segment 1", "segment 2"}, Ttl: 1},
			web.ResponseUsual{Status: "ok"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestGetSegments(t *testing.T) {
	for i, test := range []Testable{
		TestWait(1100), // + 100 ms just in case
		&TestGet{
			web.GetSegmentsBody{Id: 101},
			web.ResponseGetSegments{Status: "ok", Segments: []string{"segment 1", "hundred"}},
			web.ResponseGetSegments{Status: "ok", Segments: []string{"segment 1", "fifty-fifty", "hundred"}},
		},
		&TestGet{
			web.GetSegmentsBody{Id: 10},
			web.ResponseGetSegments{Status: "ok"}, // We didn't know of this one before, so not in any segment,
			web.ResponseGetSegments{Status: "ok"}, // not even in "hundred".
		},
		&TestGet{
			web.GetSegmentsBody{Id: 1234},
			web.ResponseGetSegments{Status: "ok", Segments: []string{"hundred"}}, // The segment perished after a second
			web.ResponseGetSegments{Status: "ok", Segments: []string{"fifty-fifty", "hundred"}},
		},
		&TestGet{
			web.GetSegmentsBody{Id: 12345},
			web.ResponseGetSegments{Status: "ok", Segments: []string{"hundred"}}, // The segment perished after a second
			web.ResponseGetSegments{Status: "ok", Segments: []string{"fifty-fifty", "hundred"}},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestHistoryPost(t *testing.T) {
	for i, test := range []Testable{
		&TestHistory{
			web.HistoryBody{1000, 1},
			web.ResponseHistory{Status: "error", Err: "bad time"},
		},
		&TestHistory{
			web.HistoryBody{2024, 13},
			web.ResponseHistory{Status: "error", Err: "bad time"},
		},
		&TestHistory{
			web.HistoryBody{2023, 6},
			web.ResponseHistory{Status: "ok", Link: "/history?year=2023&month=6"},
		},
		&TestHistory{
			web.HistoryBody{2023, 10},
			web.ResponseHistory{Status: "ok", Link: "/history?year=2023&month=10"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestHistoryGet(t *testing.T) {
	times := time.Now()
	todayYear, todayMonth := times.Year(), int(times.Month())

	table := []struct {
		year, month, status int
		wantManyLines       bool
	}{
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

func TestMain(m *testing.M) {
	go main()
	time.Sleep(200 * time.Millisecond) // Plenty of time for main() to start.
	os.Exit(m.Run())
}
