package soapkehamilan

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SoapKehamilan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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
	collection := db.Collection("soap_kehamilan")

	decoder := json.NewDecoder(r.Body)
	var dataMap map[string]interface{}
	err = decoder.Decode(&dataMap)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Invalid request body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	data := dataMap["data"].(map[string]interface{})
	data["tglDatang"] = data["soapAnc"].(map[string]interface{})["tanggal"].(string)

	idPasien, _ := data["id_pasien"].(string)
	data["id_pasien"], err = strconv.Atoi(idPasien)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Invalid id_pasien"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	//insert data to database
	_, err = collection.InsertOne(context.Background(), data)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Error inserting data to database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}
	jsonData, _ := json.Marshal(map[string]interface{}{"message": "Success"})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
