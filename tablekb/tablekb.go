package tablekb

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TableKB(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := godotenv.Load()
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": ".env not found"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error connecting to database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "Database disconnected"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
		}
	}()
	db := client.Database("mydb")
	pasien_collection := db.Collection("pasien")
	soap_collection := db.Collection("soap_kb")

	//ambil id_pasien dari query url
	id_pasien_str := r.URL.Query().Get("id_pasien")
	//cek apakah id_pasien kosong
	if id_pasien_str == "" {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "id_pasien is empty"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	//trim "[", "]" dari id_pasien_str
	id_pasien_str = id_pasien_str[1 : len(id_pasien_str)-1]
	//split id_pasien_str menjadi array dengan delimiter ","
	id_pasien_arr := strings.Split(id_pasien_str, ",")
	//convert id_pasien_arr menjadi array of int64

	var returnData []bson.M
	for _, id := range id_pasien_arr {
		id_int, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "error converting id_pasien to int64"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}
		pasien_filter := bson.M{"id_pasien": id_int}
		pasien_info := pasien_collection.FindOne(context.Background(), pasien_filter)
		if pasien_info.Err() != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "error finding pasien"})
			w.WriteHeader(http.StatusNotFound)
			w.Write(jsonData)
			return
		}

		pasien_history, err := soap_collection.Find(context.Background(), pasien_filter)
		if err != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "error finding pasien"})
			w.WriteHeader(404)
			w.Write(jsonData)
			return
		}

		var pasien_history_arr []bson.M
		if err = pasien_history.All(context.Background(), &pasien_history_arr); err != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "error decoding pasien_history"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}

		var pasienData bson.M
		if err := pasien_info.Decode(&pasienData); err != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "error decoding pasien_info"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}
		cara_kb_terakhir := pasienData["data_kb"].(bson.M)["informasi_lainnya"].(bson.M)["caraKBTerakhir"].(string)

		for _, data := range pasien_history_arr {
			tglDatang := data["tglDatang"].(string)
			tglDatang = tglDatang[:10]
			data["tglDatang"] = tglDatang
		}

		last_datang := pasien_history_arr[len(pasien_history_arr)-1]["tglDatang"].(string)

		tanggalDatetime, err := time.Parse("02-01-2006", last_datang)
		if err != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "error convert date"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}

		nama_hari_id := []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"}
		nama_bulan_id := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

		hari := nama_hari_id[tanggalDatetime.Weekday()]
		bulan := nama_bulan_id[tanggalDatetime.Month()-1]
		tahun := tanggalDatetime.Year()

		tanggal_indonesia := hari + ", " + strconv.Itoa(tanggalDatetime.Day()) + " " + bulan + " " + strconv.Itoa(tahun)

		returnData = append(returnData, bson.M{
			"id_pasien":         id_int,
			"name":              pasienData["nama_pasien"],
			"usia":              pasienData["umur"],
			"metodeKontrasepsi": cara_kb_terakhir,
			"tglDatang":         tanggal_indonesia,
			"subRows":           pasien_history_arr,
		})

		jsonData, _ := json.Marshal(map[string]interface{}{"message": "success", "data": returnData})
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}
