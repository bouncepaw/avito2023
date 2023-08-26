package swagger

import (
	"avito2023/db"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

func failWithError(err error, encoder *json.Encoder) {
	response := InlineResponse200{
		Status: "error",
		Error_: err.Error(),
	}

	err = encoder.Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func failWithGetError(err error, encoder *json.Encoder) {
	response := InlineResponse2001{
		Status: "error",
		Error_: err.Error(),
	}

	err = encoder.Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func alright(encoder *json.Encoder) {
	response := InlineResponse200{Status: "ok"}
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

	_ = encoder.Encode(InlineResponse2001{
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

	err = db.UpdateUser(context.Background(), int(body.Id), body.AddToSegments, body.RemoveFromSegments)
	if err != nil {
		failWithError(err, encoder)
		return
	}

	alright(encoder)
}

func HistoryGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func HistoryPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}