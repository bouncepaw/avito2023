package main

import (
	"avito2023/db"
	"log"
	"net/http"

	"avito2023/web"
)

func main() {
	defer db.Close()
	log.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", web.NewRouter()))
}
