package editkb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EditKb(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Connect database error"})
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

	decoder := json.NewDecoder(r.Body)
	var dataMap map[string]interface{}
	if err := decoder.Decode(&dataMap); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error decoding data from request body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}
	// log.Printf("data: %v", dataMap)

	id_pasien_float, ok := dataMap["id_pasien"].(float64)
	if !ok {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "id_pasien invalid"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}
	id_pasien := uint64(id_pasien_float)

	var filterData bson.M
	var targetPasien bson.M
	var dataPasien bson.M

	data := dataMap["data"].(map[string]interface{})
	if data == nil {
		filterData = bson.M{"id_pasien": id_pasien}
		pasien := db.Collection("pasien").FindOne(context.Background(), filterData)

		if pasien.Err() != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "id_pasien tidak ditemukan", "statusCode": 401})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		var pasienData bson.M
		if err := pasien.Decode(&pasienData); err != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "error decoding data"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}

		returnData := bson.M{
			"generalInformation": bson.M{
				"noFaskes":          pasienData["data_kb"].(bson.M)["no_faskes"],
				"noSeriKartu":       pasienData["data_kb"].(bson.M)["no_seri_kartu"],
				"tglDatang":         pasienData["tanggal_register"],
				"namaPeserta":       pasienData["nama_pasien"],
				"tglLahir":          pasienData["tanggal_lahir"],
				"usia":              pasienData["umur"],
				"namaPasangan":      pasienData["nama_pasangan"],
				"jenisPasangan":     pasienData["jenis_pasangan"],
				"pendidikanAkhir":   pasienData["pendidikan"],
				"alamat":            pasienData["alamat"],
				"pekerjaanPasangan": pasienData["pekerjaan_pasangan"],
				"statusJkn":         pasienData["data_kb"].(bson.M)["status_jkn"],
			},
			"otherInformation": pasienData["data_kb"].(bson.M)["informasi_lainnya"],
			"skrining":         pasienData["data_kb"].(bson.M)["skrining"],
			"hasil":            pasienData["data_kb"].(bson.M)["hasil"],
			"penapisanKB":      pasienData["data_kb"].(bson.M)["penapisan_kb"],
		}

		jsonData, _ := json.Marshal(map[string]interface{}{"message": "success", "data": returnData})
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
		return
	}

	targetPasien = bson.M{"id_pasien": id_pasien}
	dataPasien = bson.M{
		"tanggal_register":   data["generalInformation"].(map[string]interface{})["tglDatang"],
		"nama_pasien":        data["generalInformation"].(map[string]interface{})["namaPeserta"],
		"tanggal_lahir":      data["generalInformation"].(map[string]interface{})["tglLahir"],
		"umur":               data["generalInformation"].(map[string]interface{})["usia"],
		"nama_pasangan":      data["generalInformation"].(map[string]interface{})["namaPasangan"],
		"jenis_pasangan":     data["generalInformation"].(map[string]interface{})["jenisPasangan"],
		"pendidikan":         data["generalInformation"].(map[string]interface{})["pendidikanAkhir"],
		"alamat":             data["generalInformation"].(map[string]interface{})["alamat"],
		"pekerjaan_pasangan": data["generalInformation"].(map[string]interface{})["pekerjaanPasangan"],
		"data_kb": bson.M{
			"status_jkn":        data["generalInformation"].(map[string]interface{})["statusJkn"],
			"no_faskes":         data["generalInformation"].(map[string]interface{})["noFaskes"],
			"no_seri_kartu":     data["generalInformation"].(map[string]interface{})["noSeriKartu"],
			"informasi_lainnya": data["otherInformation"],
			"skrining":          data["skrining"],
			"hasil":             data["hasil"],
			"penapisan_kb":      data["penapisanKB"],
		},
	}

	if _, err := db.Collection("pasien").UpdateOne(context.Background(), targetPasien, bson.M{"$set": dataPasien}); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": fmt.Sprintf("error updating data for id_pasien=%d", id_pasien)})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": fmt.Sprintf("changed id_pasien=%d data", id_pasien)})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}