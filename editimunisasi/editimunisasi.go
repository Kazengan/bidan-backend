package editimunisasi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EditImunisasi(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Error loading .env file"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Connect database error"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "Database disconnected"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
		}
	}()

	db := client.Database("mydb")

	decoder := json.NewDecoder(r.Body)
	var dataMap map[string]interface{}
	if err := decoder.Decode(&dataMap); err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "error decoding data from request body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	id_pasien_str, ok := dataMap["id_pasien"].(string)
	if !ok {
		jsonData, _ := json.Marshal(map[string]string{"message": "id_pasien invalid"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	id_pasien, err := strconv.Atoi(id_pasien_str)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "invalid id_pasien"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	data, ok := dataMap["data"].(map[string]interface{})
	if !ok || len(data) == 0 {
		filterData := bson.M{"id_pasien": id_pasien}
		pasien := db.Collection("pasien").FindOne(context.Background(), filterData)

		if pasien.Err() != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "id_pasien tidak ditemukan"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		var pasienData bson.M
		if err := pasien.Decode(&pasienData); err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "error decoding data"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}

		returnData := bson.M{
			"generalInformation": bson.M{
				"nomorBayi": pasienData["nomor_bayi"],
				"nomor":     pasienData["nomor"],
				"namaBayi":  pasienData["nama_pasien"],
				"namaAyah":  pasienData["nama_ayah"],
				"usiaAyah":  pasienData["umur_ayah"],
				"namaIbu":   pasienData["nama_ibu"],
				"usiaIbu":   pasienData["umur_ibu"],
				"puskesmas": pasienData["puskesmas"],
				"bidan":     pasienData["bidan"],
				"alamat":    pasienData["alamat"],
				"desa":      pasienData["desa"],
				"kecamatan": pasienData["kecamatan"],
				"kabupaten": pasienData["kabupaten"],
				"provinsi":  pasienData["provinsi"],
			},
			"detailBayi":                  pasienData["data_imunisasi"].(bson.M)["detail_bayi"],
			"pemeriksaanNeonatus":         pasienData["data_imunisasi"].(bson.M)["pemeriksaan_neonatus"],
			"pemeriksaanNeonatusLanjutan": pasienData["data_imunisasi"].(bson.M)["pemeriksaan_neonatus_lanjutan"],
			"pemeriksaanBalita":           pasienData["data_imunisasi"].(bson.M)["pemeriksaan_balita"],
		}

		jsonData, _ := json.Marshal(map[string]interface{}{"message": "success", "data": returnData})
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
		return
	} else {
	targetPasien := bson.M{"id_pasien": id_pasien}
	dataPasien := bson.M{
		"nomor_bayi":  data["generalInformation"].(map[string]interface{})["nomorBayi"],
		"nomor":       data["generalInformation"].(map[string]interface{})["nomor"],
		"nama_pasien": data["generalInformation"].(map[string]interface{})["namaBayi"],
		"nama_ayah":   data["generalInformation"].(map[string]interface{})["namaAyah"],
		"umur_ayah":   data["generalInformation"].(map[string]interface{})["usiaAyah"],
		"nama_ibu":    data["generalInformation"].(map[string]interface{})["namaIbu"],
		"umur_ibu":    data["generalInformation"].(map[string]interface{})["usiaIbu"],
		"puskesmas":   data["generalInformation"].(map[string]interface{})["puskesmas"],
		"bidan":       data["generalInformation"].(map[string]interface{})["bidan"],
		"alamat":      data["generalInformation"].(map[string]interface{})["alamat"],
		"desa":        data["generalInformation"].(map[string]interface{})["desa"],
		"kecamatan":   data["generalInformation"].(map[string]interface{})["kecamatan"],
		"kabupaten":   data["generalInformation"].(map[string]interface{})["kabupaten"],
		"provinsi":    data["generalInformation"].(map[string]interface{})["provinsi"],
		"data_imunisasi": bson.M{
			"detail_bayi":                   data["detailBayi"],
			"pemeriksaan_neonatus":          data["pemeriksaanNeonatus"],
			"pemeriksaan_neonatus_lanjutan": data["pemeriksaanNeonatusLanjutan"],
			"pemeriksaan_balita":            data["pemeriksaanBalita"],
		},
	}

	if _, err := db.Collection("pasien").UpdateOne(context.Background(), targetPasien, bson.M{"$set": dataPasien}); err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("error updating data for id_pasien=%d", id_pasien)})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("changed id_pasien=%d data", id_pasien)})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
	}
}
