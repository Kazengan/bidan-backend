package countt

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CountHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idLayanan, err := strconv.Atoi(r.URL.Query().Get("id_layanan"))

	if err != nil {
		somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Invalid id_layanan", "statusCode": 400})
		w.Write(somethingWentWrong)
		return
	}

	if err := godotenv.Load(); err != nil {
		somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Failed to load .env file", "statusCode": 400})
		w.Write(somethingWentWrong)
		return
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Failed to connect to database", "statusCode": 400})
		w.Write(somethingWentWrong)
		return
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Database disconnected", "statusCode": 400})
			w.Write(somethingWentWrong)
			return
		}
	}()

	db := client.Database("mydb")

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
		jsonData, _ := json.Marshal(map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": 0, "lastUpdate": nil})
		w.Write(jsonData)
		return
	}

	collection := db.Collection(collectionName)
	now := time.Now().UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	filterCriteria := bson.M{dateField: bson.M{"$gte": startOfMonth.Format(time.RFC3339), "$lt": endOfMonth.Format(time.RFC3339)}}
	cursor, err := collection.Find(context.Background(), filterCriteria)
	if err != nil {
		somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Finding documents went wrong", "statusCode": 400})
		w.Write(somethingWentWrong)
		return
	}
	defer cursor.Close(context.Background())

	var lastUpdate time.Time
	var countData int64
	for cursor.Next(context.Background()) {
		countData++
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Decoding document went wrong", "statusCode": 400})
			w.Write(somethingWentWrong)
			return
		}

		var docDate time.Time
		if date, ok := doc[dateField].(string); ok {
			docDate, err = time.Parse(time.RFC3339, date)
			if err != nil {
				somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Parsing date went wrong", "statusCode": 400})
				w.Write(somethingWentWrong)
				return
			}
		}
		if docDate.After(lastUpdate) {
			lastUpdate = docDate
		}
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": countData, "lastUpdate": lastUpdate.Format(time.RFC3339)})
	w.Write(jsonData)
}
