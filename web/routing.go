package web

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

var routes = []route{
	{"Index", "GET", "/", index},
	{"CreateSegmentPost", "POST", "/create_segment", CreateSegmentPost},
	{"DeleteSegmentPost", "POST", "/delete_segment", DeleteSegmentPost},
	{"GetSegmentsPost", "POST", "/get_segments", GetSegmentsPost},
	{"UpdateUserPost", "POST", "/update_user", UpdateUserPost},
	{"HistoryGet", "GET", "/history", HistoryGet},
	{"HistoryPost", "POST", "/history", HistoryPost},
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(logger(route.HandlerFunc, route.Name))
	}

	return router
}
