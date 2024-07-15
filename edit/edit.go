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
					"noHP":              pasienData["no_hp"],
				},
				"otherInformation": pasienData["data_kb"].(bson.M)["informasi_lainnya"],
				"skrining":         pasienData["data_kb"].(bson.M)["skrining"],
				"hasil":            pasienData["data_kb"].(bson.M)["hasil"],
				"penapisanKB":      pasienData["data_kb"].(bson.M)["penapisan_kb"],
			}

		} else if id_layanan_int == 1 {
			returnData = bson.M{
				"generalInformation": bson.M{
					"agama":           pasienData["data_kehamilan"].(bson.M)["agama"],
					"pekerjaan":       pasienData["data_kehamilan"].(bson.M)["pekerjaan"],
					"desa":            pasienData["data_kehamilan"].(bson.M)["desa"],
					"kabupaten":       pasienData["data_kehamilan"].(bson.M)["kabupaten"],
					"kecamatan":       pasienData["data_kehamilan"].(bson.M)["kecamatan"],
					"provinsi":        pasienData["data_kehamilan"].(bson.M)["provinsi"],
					"rtrw":            pasienData["data_kehamilan"].(bson.M)["rtrw"],
					"noIbu":           pasienData["data_kehamilan"].(bson.M)["no_ibu"],
					"tanggalRegister": pasienData["tanggal_register"],
					"namaLengkap":     pasienData["nama_pasien"],
					"tanggalLahir":    pasienData["tanggal_lahir"],
					"umur":            pasienData["umur"],
					"namaSuami":       pasienData["nama_pasangan"],
					"pendidikan":      pasienData["pendidikan"],
					"alamatDomisili":  pasienData["alamat"],
				},
				"kunjunganNifas":                        pasienData["data_kehamilan"].(bson.M)["kunjungan_nifas"],
				"mendeteksiFaktorResikoDanResikoTinggi": pasienData["data_kehamilan"].(bson.M)["faktor_resiko_resiko_tinggi"],
				"pemeriksaanPNC":                        pasienData["data_kehamilan"].(bson.M)["pemeriksaan_pnc"],
				"persalinan":                            pasienData["data_kehamilan"].(bson.M)["persalinan"],
				"rencanaPersalinan":                     pasienData["data_kehamilan"].(bson.M)["rencana_persalinan"],
				"riwayatKehamilan":                      pasienData["data_kehamilan"].(bson.M)["riwayat_kehamilan"],
				"skriningTT":                            pasienData["data_kehamilan"].(bson.M)["skrining_tt"],
				"section2":                              pasienData["data_kehamilan"].(bson.M)["section2"],
			}

		} else if id_layanan_int == 2 {
			returnData = bson.M{
				"generalInformation": bson.M{
					"nomorBayi": pasienData["nomor_bayi"],
					"tglDatang": pasienData["tanggal_register"],
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
					"noHP":      pasienData["no_hp"],
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

		//if request POST
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

		id_layanan_int, ok := dataMap["id_layanan"].(float64)
		if !ok {
			jsonData, _ := json.Marshal(map[string]string{"message": "(POST) error id_layanan is empty"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		data, ok := dataMap["data"].(map[string]interface{})
		if !ok || len(data) == 0 {
			jsonData, _ := json.Marshal(map[string]string{"message": "(POST) error data is empty"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonData)
			return
		}

		db := client.Database("mydb")
		targetPasien := bson.M{"id_pasien": id_pasien_int}
		var dataPasien bson.M

		if id_layanan_int == 0 {
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
				"no_hp":              data["generalInformation"].(map[string]interface{})["noHP"],
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
		} else if id_layanan_int == 1 {
			dataPasien = bson.M{
				"tanggal_register": data["generalInformation"].(map[string]interface{})["tanggalRegister"],
				"nama_pasien":      data["generalInformation"].(map[string]interface{})["namaLengkap"],
				"tanggal_lahir":    data["generalInformation"].(map[string]interface{})["tanggalLahir"],
				"umur":             data["generalInformation"].(map[string]interface{})["umur"],
				"nama_pasangan":    data["generalInformation"].(map[string]interface{})["namaSuami"],
				"pendidikan":       data["generalInformation"].(map[string]interface{})["pendidikan"],
				"alamat":           data["generalInformation"].(map[string]interface{})["alamatDomisili"],
				"no_hp":            data["section2"].(map[string]interface{})["noTelp"],
				"data_kehamilan": bson.M{
					"pekerjaan":                   data["generalInformation"].(map[string]interface{})["pekerjaan"],
					"agama":                       data["generalInformation"].(map[string]interface{})["agama"],
					"desa":                        data["generalInformation"].(map[string]interface{})["desa"],
					"kabupaten":                   data["generalInformation"].(map[string]interface{})["kabupaten"],
					"kecamatan":                   data["generalInformation"].(map[string]interface{})["kecamatan"],
					"provinsi":                    data["generalInformation"].(map[string]interface{})["provinsi"],
					"rtrw":                        data["generalInformation"].(map[string]interface{})["rtrw"],
					"no_ibu":                      data["generalInformation"].(map[string]interface{})["noIbu"],
					"kunjungan_nifas":             data["kunjunganNifas"],
					"faktor_resiko_resiko_tinggi": data["mendeteksiFaktorResikoDanResikoTinggi"],
					"pemeriksaan_pnc":             data["pemeriksaanPNC"],
					"persalinan":                  data["persalinan"],
					"rencana_persalinan":          data["rencanaPersalinan"],
					"riwayat_kehamilan":           data["riwayatKehamilan"],
					"skrining_tt":                 data["skriningTT"],
					"section2":                    data["section2"],
				},
			}
		} else if id_layanan_int == 2 {
			dataPasien = bson.M{
				"nomor_bayi":       data["generalInformation"].(map[string]interface{})["nomorBayi"],
				"tanggal_register": data["generalInformation"].(map[string]interface{})["tglDatang"],
				"nomor":            data["generalInformation"].(map[string]interface{})["nomor"],
				"nama_pasien":      data["generalInformation"].(map[string]interface{})["namaBayi"],
				"nama_ayah":        data["generalInformation"].(map[string]interface{})["namaAyah"],
				"umur_ayah":        data["generalInformation"].(map[string]interface{})["usiaAyah"],
				"nama_ibu":         data["generalInformation"].(map[string]interface{})["namaIbu"],
				"umur_ibu":         data["generalInformation"].(map[string]interface{})["usiaIbu"],
				"puskesmas":        data["generalInformation"].(map[string]interface{})["puskesmas"],
				"bidan":            data["generalInformation"].(map[string]interface{})["bidan"],
				"alamat":           data["generalInformation"].(map[string]interface{})["alamat"],
				"desa":             data["generalInformation"].(map[string]interface{})["desa"],
				"kecamatan":        data["generalInformation"].(map[string]interface{})["kecamatan"],
				"kabupaten":        data["generalInformation"].(map[string]interface{})["kabupaten"],
				"provinsi":         data["generalInformation"].(map[string]interface{})["provinsi"],
				"no_hp":            data["generalInformation"].(map[string]interface{})["noHP"],
				"data_imunisasi": bson.M{
					"detail_bayi":                   data["detailBayi"],
					"pemeriksaan_neonatus":          data["pemeriksaanNeonatus"],
					"pemeriksaan_neonatus_lanjutan": data["pemeriksaanNeonatusLanjutan"],
					"pemeriksaan_balita":            data["pemeriksaanBalita"],
				},
			}
		}

		if _, err := db.Collection("pasien").UpdateOne(context.Background(), targetPasien, bson.M{"$set": dataPasien}); err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("error updating data for id_pasien=%d", id_pasien_int)})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}

		jsonData, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("success updating data for id_pasien=%d", id_pasien_int)})
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
		return
	}

}
