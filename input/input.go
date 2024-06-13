package input

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectToDatabase() (*mongo.Client, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf(".env not found")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		return nil, fmt.Errorf("error connecting to database")
	}
	return client, nil
}

func getNextPasien(db *mongo.Database) (uint64, error) {
	collection := db.Collection("pasien_counter")
	filter := bson.M{}
	update := bson.M{"$inc": bson.M{"seq_value": 1}}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedCounter bson.M
	err := collection.FindOneAndUpdate(context.Background(), filter, update, options).Decode(&updatedCounter)
	if err != nil {
		return 0, err
	}

	sequenceValue, ok := updatedCounter["seq_value"].(int64)
	if !ok {
		return 0, fmt.Errorf("sequence_value is not of type int64, sequence_value: %v, sequence_value type: %T", updatedCounter["sequence_value"], updatedCounter["sequence_value"])
	}

	return uint64(sequenceValue), nil
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, _ := json.Marshal(map[string]string{"message": message})
	w.WriteHeader(status)
	w.Write(jsonData)
}

func Input(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	client, err := connectToDatabase()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database disconnect error")
		}
	}()

	db := client.Database("mydb")
	collection := db.Collection("pasien")

	var dataMap map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&dataMap); err != nil {
		respondWithError(w, http.StatusBadRequest, "error decoding data from request body")
		return
	}

	data, ok := dataMap["data"].(map[string]interface{})
	if !ok {
		respondWithError(w, http.StatusBadRequest, "data field is required")
		return
	}

	idLayananInt, ok := dataMap["id_layanan"].(float64)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "id_layanan field is required")
		return
	}

	nextIDPasien, err := getNextPasien(db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error getting next pasien id: %v", err))
		return
	}

	var dataPasien bson.M

	switch idLayananInt {
	case 0:
		dataPasien = bson.M{
			"id_pasien":          nextIDPasien,
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

	case 1:
		dataPasien = bson.M{
			"id_pasien":        nextIDPasien,
			"tanggal_register": data["generalInformation"].(map[string]interface{})["tanggalRegister"],
			"nama_pasien":      data["generalInformation"].(map[string]interface{})["namaLengkap"],
			"tanggal_lahir":    data["generalInformation"].(map[string]interface{})["tanggalLahir"],
			"umur":             data["generalInformation"].(map[string]interface{})["umur"],
			"nama_pasangan":    data["generalInformation"].(map[string]interface{})["namaSuami"],
			"pendidikan":       data["generalInformation"].(map[string]interface{})["pendidikan"],
			"alamat":           data["generalInformation"].(map[string]interface{})["alamatDomisili"],
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
	case 2:
		dataPasien = bson.M{
			"id_pasien":   nextIDPasien,
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
			"no_hp":       data["generalInformation"].(map[string]interface{})["noHP"],
			"data_imunisasi": bson.M{
				"detail_bayi":                     data["detailBayi"],
				"pemeriksaan_neonatus":            data["pemeriksaanNeonatus"],
				"pelayanan_klinis_tumbuh_kembang": data["pelayananKlinisTumbuhKembang"],
				"pemberian_vitamin_a":             data["pemberianVitaminA"],
				"imunisasi":                       data["imunisasi"],
			},
		}
	default:
		respondWithError(w, http.StatusBadRequest, "invalid id_layanan value")
		return
	}

	_, err = collection.InsertOne(context.Background(), dataPasien)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error inserting data into database")
		return
	}

	response := map[string]interface{}{
		"message": "success",
		"id":      nextIDPasien,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error marshalling response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
