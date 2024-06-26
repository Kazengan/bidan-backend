package inputkb

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

func InputKB(w http.ResponseWriter, r *http.Request) {
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
	collection := db.Collection("pasien")

	var dataMap map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&dataMap); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error decoding data from request body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	data := dataMap["data"].(map[string]interface{})
	if data == nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "data needed"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	next_id_pasien, err := getNextPasien(db)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error() + "\naaaaaaaa"))
		return
	}

	dataPasien := bson.M{
		"id_pasien":          next_id_pasien,
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

	if _, err := collection.InsertOne(context.Background(), dataPasien); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error inserting data"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "success", "id_pasien": next_id_pasien})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}
