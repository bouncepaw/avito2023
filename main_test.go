package main

import (
	swagger "avito2023/web"
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
	swagger.ResponseUsual
}

func (tc *TestCreate) Test(idx int, t *testing.T) {
	response, ok := yesbut("create_segment", tc.CreateSegmentBody, tc.ResponseUsual)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.ResponseUsual)
	}
}

type TestDelete struct {
	swagger.DeleteSegmentBody
	swagger.ResponseUsual
}

func (tc *TestDelete) Test(idx int, t *testing.T) {
	response, ok := yesbut("delete_segment", tc.DeleteSegmentBody, tc.ResponseUsual)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.ResponseUsual)
	}
}

type TestUpdate struct {
	swagger.UpdateUserBody
	swagger.ResponseUsual
}

func (tc *TestUpdate) Test(idx int, t *testing.T) {
	response, ok := yesbut("update_user", tc.UpdateUserBody, tc.ResponseUsual)
	if !ok {
		t.Errorf("Failed test %d: got %q instead of %q", idx, response, tc.ResponseUsual)
	}
}

type TestGet struct {
	swagger.GetSegmentsBody
	possibility1 swagger.ResponseGetSegments
	possibility2 swagger.ResponseGetSegments
}

func (tc *TestGet) Test(idx int, t *testing.T) {
	var r any = post[swagger.ResponseGetSegments]("get_segments", tc.GetSegmentsBody)
	response := r.(swagger.ResponseGetSegments)

	ok1 := reflect.DeepEqual(response, tc.possibility1)
	ok2 := reflect.DeepEqual(response, tc.possibility2)
	if !(ok1 || ok2) {
		t.Errorf("Failed test %d: got %q instead of %q or %q", idx, response, tc.possibility1, tc.possibility2)
	}
}

type TestHistory struct {
	swagger.HistoryBody
	swagger.ResponseHistory
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
			swagger.CreateSegmentBody{Name: "segment 1"},
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "segment 2"},
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "segment 1"},
			swagger.ResponseUsual{Status: "error", Error_: "name taken"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "segment to delete 1"},
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "quarnishone", Percent: 123},
			swagger.ResponseUsual{Status: "error", Error_: "bad percent"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "fifty-fifty", Percent: 50}, // !
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: "hundred", Percent: 100}, // !
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestCreate{
			swagger.CreateSegmentBody{Name: ""},
			swagger.ResponseUsual{Status: "error", Error_: "name empty"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestDeleteSegment(t *testing.T) {
	for i, test := range []Testable{
		&TestDelete{
			swagger.DeleteSegmentBody{Name: "segment to delete 1"},
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestDelete{
			swagger.DeleteSegmentBody{Name: "segment to delete 1"},
			swagger.ResponseUsual{Status: "error", Error_: "segment deleted"},
		},
		&TestDelete{
			swagger.DeleteSegmentBody{Name: "quasimodo"},
			swagger.ResponseUsual{Status: "error", Error_: "name free"},
		},
	} {
		test.Test(i+1, t)
	}
}

func TestUpdateUser(t *testing.T) {
	for i, test := range []Testable{
		&TestUpdate{
			swagger.UpdateUserBody{Id: 546, AddToSegments: []string{"segment to delete 1"}},
			swagger.ResponseUsual{Status: "error", Error_: "segment deleted"},
		},
		&TestUpdate{
			swagger.UpdateUserBody{Id: 101, AddToSegments: []string{"segment 1", "segment 2"}, RemoveFromSegments: []string{}},
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestUpdate{
			swagger.UpdateUserBody{Id: 101, AddToSegments: []string{}, RemoveFromSegments: []string{"segment 2"}},
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestUpdate{
			// This shall perish in a second
			swagger.UpdateUserBody{Id: 1234, AddToSegments: []string{"segment 1"}, Ttl: 1},
			swagger.ResponseUsual{Status: "ok"},
		},
		&TestGet{
			swagger.GetSegmentsBody{Id: 1234},
			swagger.ResponseGetSegments{Status: "ok", Segments: []string{"segment 1", "hundred"}}, // The segment has not yet perished
			swagger.ResponseGetSegments{Status: "ok", Segments: []string{"segment 1", "fifty-fifty", "hundred"}},
		},
		&TestUpdate{
			// This shall perish in a second
			swagger.UpdateUserBody{Id: 12345, AddToSegments: []string{"segment 1", "segment 2"}, Ttl: 1},
			swagger.ResponseUsual{Status: "ok"},
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
			swagger.ResponseGetSegments{Status: "ok", Segments: []string{"segment 1", "hundred"}},
			swagger.ResponseGetSegments{Status: "ok", Segments: []string{"segment 1", "fifty-fifty", "hundred"}},
		},
		&TestGet{
			swagger.GetSegmentsBody{Id: 10},
			swagger.ResponseGetSegments{Status: "ok"}, // We didn't know of this one before, so not in any segment,
			swagger.ResponseGetSegments{Status: "ok"}, // not even in "hundred".
		},
		&TestGet{
			swagger.GetSegmentsBody{Id: 1234},
			swagger.ResponseGetSegments{Status: "ok", Segments: []string{"hundred"}}, // The segment perished after a second
			swagger.ResponseGetSegments{Status: "ok", Segments: []string{"fifty-fifty", "hundred"}},
		},
		&TestGet{
			swagger.GetSegmentsBody{Id: 12345},
			swagger.ResponseGetSegments{Status: "ok", Segments: []string{"hundred"}}, // The segment perished after a second
			swagger.ResponseGetSegments{Status: "ok", Segments: []string{"fifty-fifty", "hundred"}},
		},
	} {
		test.Test(i+1, t)
	}
}

// TODO: Investigate

func TestHistoryPost(t *testing.T) {
	for i, test := range []Testable{
		&TestHistory{
			swagger.HistoryBody{1000, 1},
			swagger.ResponseHistory{Status: "error", Error_: "bad time"},
		},
		&TestHistory{
			swagger.HistoryBody{2024, 13},
			swagger.ResponseHistory{Status: "error", Error_: "bad time"},
		},
		&TestHistory{
			swagger.HistoryBody{2023, 6},
			swagger.ResponseHistory{Status: "ok", Link: "/history?year=2023&month=6"},
		},
		&TestHistory{
			swagger.HistoryBody{2023, 10},
			swagger.ResponseHistory{Status: "ok", Link: "/history?year=2023&month=10"},
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

func TestMain(m *testing.M) {
	go main()
	time.Sleep(200 * time.Millisecond) // Plenty of time for main() to start.
	os.Exit(m.Run())
}
