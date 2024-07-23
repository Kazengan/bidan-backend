package count

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

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

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	return client, nil
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

func CountHandler(w http.ResponseWriter, r *http.Request) {
	client, err := connectToDatabase()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to database")
		return
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database disconnect error")
			return
		}
	}()

	db := client.Database("mydb")

	// Retrieve optional id_layanan parameter
	idLayananStr := r.URL.Query().Get("id_layanan")
	var idLayanan int
	if idLayananStr != "" {
		idLayanan, err = strconv.Atoi(idLayananStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid id_layanan")
			return
		}
	} else {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": 0, "lastUpdate": nil})
		return
	}

	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())
	strTanggal := fmt.Sprintf("%04d-%02d", currentYear, currentMonth)

	var collectionName string
	var dateField string

	switch idLayanan {
	case 0:
		collectionName = "soap_kb"
		dateField = "tglDatang"
	case 1:
		collectionName = "soap_kehamilan"
		dateField = "tglDatang"
	case 2:
		collectionName = "soap_imunisasi"
		dateField = "tglDatang"
	default:
		respondWithJSON(w, http.StatusOK, map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": 0, "lastUpdate": nil})
		return
	}

	collection := db.Collection(collectionName)
	filterCriteria := bson.M{dateField: bson.M{"$regex": "^" + strTanggal}}
	cursor, err := collection.Find(context.Background(), filterCriteria)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error finding documents")
		return
	}
	defer cursor.Close(context.Background())

	var lastUpdate string
	var countData int64
	for cursor.Next(context.Background()) {
		countData++
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error decoding document")
			return
		}

		if date, ok := doc[dateField].(string); ok {
			lastUpdate = date
		}
	}

	if countData == 0 {
		lastUpdate = fmt.Sprintf("%s-01T00:00:00Z", strTanggal)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": countData, "lastUpdate": lastUpdate, "strTanggal": strTanggal})
}
