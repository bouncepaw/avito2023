package web

import (
	"avito2023/db"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func index(w http.ResponseWriter, rq *http.Request) {
	_, _ = fmt.Fprintf(w, "Avito!")
}

func failWithError(err error, encoder *json.Encoder) {
	response := ResponseUsual{
		Status: "error",
		Error_: err.Error(),
	}

	err = encoder.Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func failWithGetError(err error, encoder *json.Encoder) {
	response := ResponseGetSegments{
		Status: "error",
		Error_: err.Error(),
	}

	err = encoder.Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func failWithHistoryError(err error, encoder *json.Encoder) {
	response := ResponseHistory{
		Status: "error",
		Error_: err.Error(),
	}

	err = encoder.Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func alright(encoder *json.Encoder) {
	response := ResponseUsual{Status: "ok"}
	err := encoder.Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateSegmentPost(w http.ResponseWriter, rq *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	var (
		body    CreateSegmentBody
		decoder = json.NewDecoder(rq.Body)
		encoder = json.NewEncoder(w)
	)

	err := decoder.Decode(&body)
	if err != nil {
		failWithError(err, encoder)
		return
	}

	err = db.CreateSegment(context.Background(), body.Name, uint(body.Percent))
	if err != nil {
		failWithError(err, encoder)
		return
	}

	alright(encoder)
}

func DeleteSegmentPost(w http.ResponseWriter, rq *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	var (
		body    DeleteSegmentBody
		decoder = json.NewDecoder(rq.Body)
		encoder = json.NewEncoder(w)
	)

	err := decoder.Decode(&body)
	if err != nil {
		failWithError(err, encoder)
		return
	}

	err = db.DeleteSegment(context.Background(), body.Name)
	if err != nil {
		failWithError(err, encoder)
		return
	}

	alright(encoder)
}

func GetSegmentsPost(w http.ResponseWriter, rq *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	var (
		body    GetSegmentsBody
		decoder = json.NewDecoder(rq.Body)
		encoder = json.NewEncoder(w)
	)

	err := decoder.Decode(&body)
	if err != nil {
		failWithGetError(err, encoder)
		return
	}

	segments, err := db.GetSegments(context.Background(), int(body.Id))
	if err != nil {
		failWithGetError(err, encoder)
		return
	}

	_ = encoder.Encode(ResponseGetSegments{
		Status:   "ok",
		Segments: segments,
	})
}

func UpdateUserPost(w http.ResponseWriter, rq *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	var (
		body    UpdateUserBody
		decoder = json.NewDecoder(rq.Body)
		encoder = json.NewEncoder(w)
	)

	err := decoder.Decode(&body)
	if err != nil {
		failWithError(err, encoder)
		return
	}

	err = db.UpdateUser(context.Background(), int(body.Id), body.AddToSegments, body.RemoveFromSegments, int(body.Ttl))
	if err != nil {
		failWithError(err, encoder)
		return
	}

	alright(encoder)
}

func showErrorStatus(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintln(w, fmt.Sprintf("%d", statusCode))
}

func HistoryGet(w http.ResponseWriter, rq *http.Request) {
	var (
		year, errYear   = strconv.Atoi(rq.FormValue("year"))
		month, errMonth = strconv.Atoi(rq.FormValue("month"))
	)

	if errYear != nil || errMonth != nil || year < 2023 || month < 1 || month > 12 {
		showErrorStatus(w, http.StatusNotFound)
		return
	}

	csv, err := db.GetHistory(context.Background(), year, month)
	if err != nil {
		log.Println(err)
		showErrorStatus(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, csv)
}

func HistoryPost(w http.ResponseWriter, rq *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	var (
		body    HistoryBody
		decoder = json.NewDecoder(rq.Body)
		encoder = json.NewEncoder(w)
	)

	err := decoder.Decode(&body)
	if err != nil {
		failWithHistoryError(err, encoder)
		return
	}

	if body.Year < 2023 || body.Month < 1 || body.Month > 12 {
		failWithHistoryError(errors.New("bad time"), encoder)
		return
	}

	_ = encoder.Encode(ResponseHistory{
		Status: "ok",
		Link:   fmt.Sprintf("/history?year=%d&month=%d", body.Year, body.Month),
	})
}
