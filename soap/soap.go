package soap

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Soap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": ".env not found"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error connecting to database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}
	defer client.Disconnect(context.Background())

	// Decode request body
	decoder := json.NewDecoder(r.Body)
	var dataMap map[interface{}]interface{}
	if err := decoder.Decode(&dataMap); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Invalid request body", "error": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	// Handle specific services based on id_layanan
	idLayananInt, ok := dataMap["id_layanan"].(float64)
	if !ok {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Invalid id_layanan"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	var collection *mongo.Collection
	switch idLayananInt {
	case 1:
		collection = client.Database("mydb").Collection("soap_kb")
	case 2:
		collection = client.Database("mydb").Collection("soap_kehamilan")
	case 3:
		collection = client.Database("mydb").Collection("soap_imunisasi")
	default:
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Service under construction"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	// Insert data to MongoDB
	if _, err := collection.InsertOne(context.Background(), dataMap["data"]); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Error inserting data to database", "error": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "Success"})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
