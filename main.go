package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Kazengan/bidan-backend/getpasien"
	"github.com/Kazengan/bidan-backend/bidanlogin"
	"github.com/Kazengan/bidan-backend/count"
	"github.com/Kazengan/bidan-backend/editkb"
	"github.com/Kazengan/bidan-backend/findpasien"
)

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	http.HandleFunc("/api/bidanlogin", bidanlogin.LoginHandler)
	http.HandleFunc("/api/getpasien", getpasien.GetPasien)
	http.HandleFunc("/api/count", count.CountHandler)
	http.HandleFunc("/api/editkb", editkb.EditKb)
	http.HandleFunc("/api/findpasien", findpasien.PasienPerLayanan)
	http.HandleFunc("/api/getpasien", getpasien.GetPasien)
	log.Printf("Listening on %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
