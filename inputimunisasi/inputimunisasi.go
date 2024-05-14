package inputimunisasi

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

// make a to check dotenv file and return error if err
func checkEnv() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("error loading .env file")
	}
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		return "", fmt.Errorf("MONGODB_URI environment variable is not set")
	}
	return mongoURI, nil
}

func checkDB(mongoURI string) (*mongo.Database, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		defer func(client *mongo.Client) {
			if err := client.Disconnect(context.Background()); err != nil {
				fmt.Printf("Error disconnecting from MongoDB: %v", err)
			}
		}(client)
		return nil, fmt.Errorf("error connecting to MongoDB: %v", err)
	}
	db := client.Database("mydb")
	return db, nil
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

func getNextNoBayi(db *mongo.Database) (uint64, error) {
	collection := db.Collection("bayi_counter")
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

func InputImunisasi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mongoURI, err := checkEnv()

	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	db, err := checkDB(mongoURI)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	pasien_collection := db.Collection("pasien")

	var dataMap map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&dataMap); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "data needed in request body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	data := dataMap["data"].(map[string]interface{})
	if data == nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "data needed in request body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	nextPasienID, err := getNextPasien(db)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	nextBayiID, err := getNextNoBayi(db)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	dataPasien := bson.M{
		"id_pasien":   nextPasienID,
		"no_bayi":     nextBayiID,
		"nama_pasien": data["generalInformation"].(map[string]string)["namaBayi"],
		"nama_ayah":   data["generalInformation"].(map[string]string)["namaSuami"],
		"umur_ayah":   data["generalInformation"].(map[string]interface{})["usiaSuami"],
		"nama_ibu":    data["generalInformation"].(map[string]string)["namaIbu"],
		"umur_ibu":    data["generalInformation"].(map[string]interface{})["usiaIbu"],
		"puskesmas":   data["generalInformation"].(map[string]string)["puskesmas"],
		"bidan":       data["generalInformation"].(map[string]string)["bidan"],
		"alamat":      data["generalInformation"].(map[string]string)["alamat"],
		"desa":        data["generalInformation"].(map[string]string)["desa"],
		"kecamatan":   data["generalInformation"].(map[string]string)["kecamatan"],
		"kabupaten":   data["generalInformation"].(map[string]string)["kabupaten"],
		"provinsi":    data["generalInformation"].(map[string]string)["provinsi"],
		"data_imunisasi": bson.M{
			"detail_bayi":                   data["detailBayi"],
			"pemeriksaan_neonatus":          data["pemeriksaanNeonatus"],
			"pemeriksaan_neonatus_lanjutan": data["pemeriksaanNeonatusLanjutan"],
			"pemeriksaan_balita":            data["pemeriksaanBalita"],
		},
	}

	_, err = pasien_collection.InsertOne(context.Background(), dataPasien)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "failed to insert data"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "data inserted successfully", "id_pasien": nextPasienID, "no_bayi": nextBayiID})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
