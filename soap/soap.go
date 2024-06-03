package soap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
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
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return client, nil
}

func insertData(w http.ResponseWriter, collectionName string, data map[string]interface{}) {
	client, err := connectToDatabase()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer client.Disconnect(context.Background())

	db := client.Database("mydb")
	collection := db.Collection(collectionName)

	idPasien, err := strconv.Atoi(data["id_pasien"].(string))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid id_pasien")
		return
	}
	data["id_pasien"] = idPasien

	if collectionName == "soap_kehamilan" {
		data["tglDatang"] = data["soapAnc"].(map[string]interface{})["tanggal"].(string)
	}

	if _, err := collection.InsertOne(context.Background(), data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error inserting data to database")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"message": "Success"})
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, _ := json.Marshal(map[string]string{"message": message})
	w.WriteHeader(status)
	w.Write(jsonData)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, _ := json.Marshal(payload)
	w.WriteHeader(status)
	w.Write(jsonData)
}

func Soap(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var dataMap map[string]interface{}
	err := decoder.Decode(&dataMap)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Invalid request body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	idLayanan, ok := dataMap["id_layanan"].(float64)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Invalid or missing id_layanan")
		return
	}

	data, ok := dataMap["data"].(map[string]interface{})
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing 'data' field in request body")
		return
	}

	switch int(idLayanan) {
	case 0:
		insertData(w, "soap_kb", data)
	case 1:
		insertData(w, "soap_kehamilan", data)
	case 2:
		insertData(w, "soap_imunisasi", data)
	default:
		respondWithError(w, http.StatusBadRequest, "Invalid id_layanan value")
	}
}