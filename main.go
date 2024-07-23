package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Kazengan/bidan-backend/allsoap"
	"github.com/Kazengan/bidan-backend/bidanlogin"
	"github.com/Kazengan/bidan-backend/chart"
	"github.com/Kazengan/bidan-backend/chartt"
	"github.com/Kazengan/bidan-backend/count"
	"github.com/Kazengan/bidan-backend/countanually"
	"github.com/Kazengan/bidan-backend/countt"
	"github.com/Kazengan/bidan-backend/delete"
	"github.com/Kazengan/bidan-backend/deletebidan"
	"github.com/Kazengan/bidan-backend/edit"
	"github.com/Kazengan/bidan-backend/editimunisasi"
	"github.com/Kazengan/bidan-backend/editkb"
	"github.com/Kazengan/bidan-backend/export"
	"github.com/Kazengan/bidan-backend/findpasien"
	"github.com/Kazengan/bidan-backend/getbidan"
	"github.com/Kazengan/bidan-backend/getpasien"
	"github.com/Kazengan/bidan-backend/getreservasi"
	"github.com/Kazengan/bidan-backend/helper"
	"github.com/Kazengan/bidan-backend/input"
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
	"github.com/Kazengan/bidan-backend/tablekehamilan"
)

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	http.HandleFunc("/api/allsoap", allsoap.Allsoap)
	http.HandleFunc("/api/bidanlogin", bidanlogin.LoginHandler)
	http.HandleFunc("/api/getpasien", getpasien.GetPasien)
	http.HandleFunc("/api/getreservasi", getreservasi.GetReservasi)
	http.HandleFunc("/api/count", count.CountHandler)
	http.HandleFunc("/api/countanually", countanually.CountHandler)
	http.HandleFunc("/api/countt", countt.CountHandler)
	http.HandleFunc("/api/chart", chart.Chart)
	http.HandleFunc("/api/chartt", chartt.Chartt)
	http.HandleFunc("/api/delete", delete.Delete)
	http.HandleFunc("/api/editkb", editkb.EditKb)
	http.HandleFunc("/api/editimunisasi", editimunisasi.EditImunisasi)
	http.HandleFunc("/api/edit", edit.Edit)
	http.HandleFunc("/api/findpasien", findpasien.PasienPerLayanan)
	http.HandleFunc("/api/inputkb", inputkb.InputKB)
	http.HandleFunc("/api/input", input.Input)
	http.HandleFunc("/api/soap", soap.Soap)

	http.HandleFunc("/api/soapkb", soapkb.SoapKB)
	http.HandleFunc("/api/soapimunisasi", soapimunisasi.SoapImunisasi)
	http.HandleFunc("/api/soapkehamilan", soapkehamilan.SoapKehamilan)
	http.HandleFunc("/api/tablekb", tablekb.TableKB)
	http.HandleFunc("/api/table", table.Table)
	http.HandleFunc("/api/tableimunisasi", tableimunisasi.TableImunisasi)
	http.HandleFunc("/api/tablekehamilan", tablekehamilan.TableKehamilan)
	http.HandleFunc("/api/inputkehamilan", inputkehamilan.InputKehamilan)
	http.HandleFunc("/api/inputimunisasi", inputimunisasi.InputImunisasi)
	http.HandleFunc("/api/getbidan", getallbidan.GetAllBidan)
	http.HandleFunc("/api/deletebidan", deletebidan.DeleteBidan)
	http.HandleFunc("/api/registbidan", registbidan.RegistBidan)
	http.HandleFunc("/api/registpasien", registpasien.RegistPasien)
	http.HandleFunc("/api/helper", helper.Helper)
	http.HandleFunc("/api/export", export.Export)

	log.Printf("Listening on %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
