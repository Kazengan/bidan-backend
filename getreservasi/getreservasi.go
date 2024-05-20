package getreservasi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetReservasi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	if r.Method != "GET" {
		jsonData, _ := json.Marshal(map[string]string{"message": "Method not allowed"})
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(jsonData)
		return
	}

	tanggal := r.URL.Query().Get("tanggal")
	if tanggal == "" {
		jsonData, _ := json.Marshal(map[string]string{"message": "tanggal is needed"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	id_layanan := r.URL.Query().Get("id_layanan")
	if id_layanan == "" {
		jsonData, _ := json.Marshal(map[string]string{"message": "id_layanan is needed"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	id_layanan_int, err := strconv.Atoi(id_layanan)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "id_layanan must be a number"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	db := client.Database("mydb")
	collection := db.Collection("reservasi_layanan")

	filter := bson.M{
		"hariReservasi": tanggal,
		"id_layanan":    id_layanan_int,
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Error finding data"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Error decoding data"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	//result  is empty return empty array
	if len(results) == 0 {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Success", "data": []bson.M{}})
		w.Write(jsonData)
		return
	}

	//if not empty return the data
	jsonData, _ := json.Marshal(map[string]interface{}{"message": "Success", "data": results})
	w.Write(jsonData)
	w.WriteHeader(http.StatusOK)

}