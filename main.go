package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Kazengan/bidan-backend/bidanlogin"
	"github.com/Kazengan/bidan-backend/chart"
	"github.com/Kazengan/bidan-backend/count"
	"github.com/Kazengan/bidan-backend/deletebidan"
	"github.com/Kazengan/bidan-backend/edit"
	"github.com/Kazengan/bidan-backend/editimunisasi"
	"github.com/Kazengan/bidan-backend/editkb"
	"github.com/Kazengan/bidan-backend/findpasien"
	"github.com/Kazengan/bidan-backend/getallbidan"
	"github.com/Kazengan/bidan-backend/getpasien"
	"github.com/Kazengan/bidan-backend/getreservasi"
	"github.com/Kazengan/bidan-backend/helper"
	"github.com/Kazengan/bidan-backend/inputimunisasi"
	"github.com/Kazengan/bidan-backend/inputkb"
	"github.com/Kazengan/bidan-backend/inputkehamilan"
	"github.com/Kazengan/bidan-backend/registbidan"
	"github.com/Kazengan/bidan-backend/registpasien"
	"github.com/Kazengan/bidan-backend/soap"
	"github.com/Kazengan/bidan-backend/soapimunisasi"
	"github.com/Kazengan/bidan-backend/soapkb"
	"github.com/Kazengan/bidan-backend/soapkehamilan"
	"github.com/Kazengan/bidan-backend/table"
	"github.com/Kazengan/bidan-backend/tableimunisasi"
	"github.com/Kazengan/bidan-backend/tablekb"
)

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	http.HandleFunc("/api/bidanlogin", bidanlogin.LoginHandler)
	http.HandleFunc("/api/getpasien", getpasien.GetPasien)
	http.HandleFunc("/api/getreservasi", getreservasi.GetReservasi)
	http.HandleFunc("/api/count", count.CountHandler)
	http.HandleFunc("/api/chart", chart.Chart)
	http.HandleFunc("/api/editkb", editkb.EditKb)
	http.HandleFunc("/api/editimunisasi", editimunisasi.EditImunisasi)
	http.HandleFunc("/api/edit", edit.Edit)
	http.HandleFunc("/api/findpasien", findpasien.PasienPerLayanan)
	http.HandleFunc("/api/inputkb", inputkb.InputKB)
	http.HandleFunc("/api/soapkb", soap.Soap)
	http.HandleFunc("/api/soapkb", soapkb.SoapKB)
	http.HandleFunc("/api/soapimunisasi", soapimunisasi.SoapImunisasi)
	http.HandleFunc("/api/soapkehamilan", soapkehamilan.SoapKehamilan)
	http.HandleFunc("/api/tablekb", tablekb.TableKB)
	http.HandleFunc("/api/table", table.Table)
	http.HandleFunc("/api/tableimunisasi", tableimunisasi.TableImunisasi)
	http.HandleFunc("/api/inputkehamilan", inputkehamilan.InputKehamilan)
	http.HandleFunc("/api/inputimunisasi", inputimunisasi.InputImunisasi)
	http.HandleFunc("/api/getallbidan", getallbidan.GetAllBidan)
	http.HandleFunc("/api/deletebidan", deletebidan.DeleteBidan)
	http.HandleFunc("/api/registbidan", registbidan.RegistBidan)
	http.HandleFunc("/api/registpasien", registpasien.RegistPasien)
	http.HandleFunc("/api/helper	", helper.Helper)

	log.Printf("Listening on %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
