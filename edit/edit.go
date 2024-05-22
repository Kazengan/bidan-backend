package edit

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

func Edit(w http.ResponseWriter, r *http.Request) {
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
			jsonData, _ := json.Marshal(map[string]string{"message": "Error discoifnnecting from database"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}
	}()

	if r.Method == "GET" {
		id_pasien_str := r.URL.Query().Get("id_pasien")
		if id_pasien_str == "" {
			jsonData, _ := json.Marshal(map[string]string{"message": "(GET) id_pasien invalid"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		id_pasien_int, err := strconv.Atoi(id_pasien_str)
		if err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "(GET) error converting id_pasien to integer"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		id_layanan_str := r.URL.Query().Get("id_layanan")
		if id_layanan_str == "" {
			jsonData, _ := json.Marshal(map[string]string{"message": "(GET) id_layanan invalid"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		id_layanan_int, err := strconv.Atoi(id_layanan_str)
		if err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "(GET) error converting id_layanan to integer"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		db := client.Database("mydb")
		filterData := bson.M{"id_pasien": id_pasien_int}
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

		var returnData bson.M
		if id_layanan_int == 0 {
			returnData = bson.M{
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

		} else if id_layanan_int == 1 {
			jsonData, _ := json.Marshal(map[string]string{"message": "Under construction"})
			w.WriteHeader(http.StatusOK)
			w.Write(jsonData)
			return

		} else if id_layanan_int == 2 {
			returnData = bson.M{
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
		}

		jsonData, _ := json.Marshal(map[string]interface{}{"message": "success", "data": returnData})
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
		return

	} else {
		decoder := json.NewDecoder(r.Body)
		var dataMap map[string]interface{}
		if err := decoder.Decode(&dataMap); err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "(POST) error decoding data from request body"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		id_pasien_str, ok := dataMap["id_pasien"].(string)
		if !ok {
			jsonData, _ := json.Marshal(map[string]string{"message": "(POST) error id_pasien is empty"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		id_pasien_int, err := strconv.Atoi(id_pasien_str)
		if err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "(POST) error converting id_pasien to integer"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		id_layanan_str, ok := dataMap["id_layanan"].(string)
		if !ok {
			jsonData, _ := json.Marshal(map[string]string{"message": "(POST) error id_layanan is empty"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		id_layanan_int, err := strconv.Atoi(id_layanan_str)
		if err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "(POST) error converting id_layanan to integer"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		db := client.Database("mydb")
		targetPasien := bson.M{"id_pasien": id_pasien_int}
		var dataPasien bson.M

		if id_layanan_int == 0 {
			dataPasien = bson.M{
				"tanggal_register":   dataMap["generalInformation"].(map[string]interface{})["tglDatang"],
				"nama_pasien":        dataMap["generalInformation"].(map[string]interface{})["namaPeserta"],
				"tanggal_lahir":      dataMap["generalInformation"].(map[string]interface{})["tglLahir"],
				"umur":               dataMap["generalInformation"].(map[string]interface{})["usia"],
				"nama_pasangan":      dataMap["generalInformation"].(map[string]interface{})["namaPasangan"],
				"jenis_pasangan":     dataMap["generalInformation"].(map[string]interface{})["jenisPasangan"],
				"pendidikan":         dataMap["generalInformation"].(map[string]interface{})["pendidikanAkhir"],
				"alamat":             dataMap["generalInformation"].(map[string]interface{})["alamat"],
				"pekerjaan_pasangan": dataMap["generalInformation"].(map[string]interface{})["pekerjaanPasangan"],
				"data_kb": bson.M{
					"status_jkn":        dataMap["generalInformation"].(map[string]interface{})["statusJkn"],
					"no_faskes":         dataMap["generalInformation"].(map[string]interface{})["noFaskes"],
					"no_seri_kartu":     dataMap["generalInformation"].(map[string]interface{})["noSeriKartu"],
					"informasi_lainnya": dataMap["otherInformation"],
					"skrining":          dataMap["skrining"],
					"hasil":             dataMap["hasil"],
					"penapisan_kb":      dataMap["penapisanKB"],
				},
			}
		} else if id_layanan_int == 1 {
			jsonData, _ := json.Marshal(map[string]string{"message": "Under construction"})
			w.WriteHeader(http.StatusOK)
			w.Write(jsonData)
			return
		} else if id_layanan_int == 2 {
			dataPasien = bson.M{
				"nomor_bayi":  dataMap["generalInformation"].(map[string]interface{})["nomorBayi"],
				"nomor":       dataMap["generalInformation"].(map[string]interface{})["nomor"],
				"nama_pasien": dataMap["generalInformation"].(map[string]interface{})["namaBayi"],
				"nama_ayah":   dataMap["generalInformation"].(map[string]interface{})["namaAyah"],
				"umur_ayah":   dataMap["generalInformation"].(map[string]interface{})["usiaAyah"],
				"nama_ibu":    dataMap["generalInformation"].(map[string]interface{})["namaIbu"],
				"umur_ibu":    dataMap["generalInformation"].(map[string]interface{})["usiaIbu"],
				"puskesmas":   dataMap["generalInformation"].(map[string]interface{})["puskesmas"],
				"bidan":       dataMap["generalInformation"].(map[string]interface{})["bidan"],
				"alamat":      dataMap["generalInformation"].(map[string]interface{})["alamat"],
				"desa":        dataMap["generalInformation"].(map[string]interface{})["desa"],
				"kecamatan":   dataMap["generalInformation"].(map[string]interface{})["kecamatan"],
				"kabupaten":   dataMap["generalInformation"].(map[string]interface{})["kabupaten"],
				"provinsi":    dataMap["generalInformation"].(map[string]interface{})["provinsi"],
				"data_imunisasi": bson.M{
					"detail_bayi":                   dataMap["detailBayi"],
					"pemeriksaan_neonatus":          dataMap["pemeriksaanNeonatus"],
					"pemeriksaan_neonatus_lanjutan": dataMap["pemeriksaanNeonatusLanjutan"],
					"pemeriksaan_balita":            dataMap["pemeriksaanBalita"],
				},
			}
		}

		if _, err := db.Collection("pasien").UpdateOne(context.Background(), targetPasien, bson.M{"$set": dataPasien}); err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("error updating data for id_pasien=%d", id_pasien_int)})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}

	}

}
