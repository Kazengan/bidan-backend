package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Helper(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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

	decoder := json.NewDecoder(r.Body)
	var dataMap map[string]interface{}
	if err := decoder.Decode(&dataMap); err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "error decoding data from request body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}
	// log.Printf("data: %v", dataMap)

	id_pasien_str, ok := dataMap["id_pasien"].(string)
	if !ok {
		jsonData, _ := json.Marshal(map[string]string{"message": "id_pasien invalid"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	id_pasien_int, err := strconv.Atoi(id_pasien_str)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "error converting id_pasien to integer"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	data, ok := dataMap["data"].(map[string]interface{})

	//check if data is not empty or present in the request
	if !ok || len(data) == 0{
		// db := client.Database("mydb")
		// coll := db.Collection("pasien")
		// filter := bson.M{"id_pasien": id_pasien_int}

		// pasien := coll.FindOne(context.Background(), filter)
		// if pasien.Err() != nil {
		// 	jsonData, _ := json.Marshal(map[string]string{"message": "id_pasien tidak ditemukan"})
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write(jsonData)
		// 	return
		// }

		// var pasienData bson.M
		// if err := pasien.Decode(&pasienData); err != nil {
		// 	jsonData, _ := json.Marshal(map[string]string{"message": "error decoding data"})
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	w.Write(jsonData)
		// 	return
		// }

		jsonData, _ := json.Marshal(map[string]interface{}{"message": "success", "id_pasien": id_pasien_int, "id_pasien_type": fmt.Sprintf("%T", id_pasien_int), "data": data})
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)

	} else {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "success", "id_pasien": id_pasien_int, "id_pasien_type": fmt.Sprintf("%T", id_pasien_int), "data": "empty data"})
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}
