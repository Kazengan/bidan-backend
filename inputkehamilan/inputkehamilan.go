package inputkehamilan

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

func getNextPasien(db *mongo.Database) (uint64, error) {
	collection := db.Collection("pasien_counter")
	filter := bson.M{}
	update := bson.M{"$inc": bson.M{"seq_value": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedCounter bson.M
	err := collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedCounter)
	if err != nil {
		return 0, err
	}

	sequenceValue, ok := updatedCounter["seq_value"].(int64)
	if !ok {
		return 0, fmt.Errorf("sequence_value is not of type int64, sequence_value: %v, sequence_value type: %T", updatedCounter["seq_value"], updatedCounter["seq_value"])
	}

	return uint64(sequenceValue), nil
}

func InputKehamilan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := godotenv.Load(); err != nil {
		http.Error(w, `{"message": ".env not found"}`, http.StatusInternalServerError)
		return
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		http.Error(w, `{"message": "error connecting to database"}`, http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			http.Error(w, `{"message": "error disconnecting from database"}`, http.StatusInternalServerError)
		}
	}()

	db := client.Database("mydb")
	collection := db.Collection("pasien")

	var dataMap map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&dataMap); err != nil {
		http.Error(w, `{"message": "error decoding data from request body"}`, http.StatusBadRequest)
		return
	}

	data, ok := dataMap["data"].(map[string]interface{})
	if !ok {
		http.Error(w, `{"message": "data field is required"}`, http.StatusBadRequest)
		return
	}

	nextIDPasien, err := getNextPasien(db)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "error getting next pasien id: %v"}`, err), http.StatusInternalServerError)
		return
	}

	dataPasien := bson.M{
		"id_pasien":          nextIDPasien,
		"tanggal_register":   data["generalInformation"].(map[string]interface{})["tanggalRegister"],
		"nama_pasien":        data["generalInformation"].(map[string]interface{})["namaLengkap"],
		"tanggal_lahir":      data["generalInformation"].(map[string]interface{})["tanggalLahir"],
		"umur":               data["generalInformation"].(map[string]interface{})["umur"],
		"nama_pasangan":      data["generalInformation"].(map[string]interface{})["namaSuami"],
		"jenis_pasangan":     data["generalInformation"].(map[string]interface{})["jenisPasangan"],
		"pendidikan":         data["generalInformation"].(map[string]interface{})["pendidikan"],
		"alamat":             data["generalInformation"].(map[string]interface{})["alamatDomisili"],
		"pekerjaan_pasangan": data["generalInformation"].(map[string]interface{})["pekerjaanPasangan"],
		"data_kb": bson.M{
			"pekerjaan":                   data["generalInformation"].(map[string]interface{})["pekerjaan"],
			"agama":                       data["generalInformation"].(map[string]interface{})["agama"],
			"desa":                        data["generalInformation"].(map[string]interface{})["desa"],
			"kabupaten":                   data["generalInformation"].(map[string]interface{})["kabupaten"],
			"kecamatan":                   data["generalInformation"].(map[string]interface{})["kecamatan"],
			"provinsi":                    data["generalInformation"].(map[string]interface{})["provinsi"],
			"rtrw":                        data["generalInformation"].(map[string]interface{})["rtrw"],
			"no_ibu":                      data["generalInformation"].(map[string]interface{})["noIbu"],
			"status_jkn":                  data["generalInformation"].(map[string]interface{})["statusJkn"],
			"no_faskes":                   data["generalInformation"].(map[string]interface{})["noFaskes"],
			"no_seri_kartu":               data["generalInformation"].(map[string]interface{})["noSeriKartu"],
			"kunjungan_nifas":             data["kunjunganNifas"],
			"faktor_resiko_resiko_tinggi": data["mendeteksiFaktorResikoDanResikoTinggi"],
			"pemeriksaan_pnc":             data["pemeriksaanPNC"],
			"persalinan":                  data["persalinan"],
			"rencana_persalinan":          data["rencanaPersalinan"],
			"riwayat_kehamilan":           data["riwayatKehamilan"],
			"screeningTT":                 data["screeningTT"],
			"section2":                    data["section2"],
		},
	}

	if _, err := collection.InsertOne(context.Background(), dataPasien); err != nil {
		http.Error(w, `{"message": "error inserting data"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{"message": "success", "id_pasien": nextIDPasien}
	jsonData, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
