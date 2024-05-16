package count

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
		somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Failed to load.env file", "statusCode": 400})
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

	switch idLayanan {
	case 0:
		collection := db.Collection("soap_kb")
		now := time.Now().UTC()

		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		formattedStartOfMonth := startOfMonth.Format("2006-01-02T15:04:05.000Z")

		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		formattedEndOfMonth := endOfMonth.Format("2006-01-02T15:04:05.000Z") // Format end of month               // Log end of month for debugging

		filterCriteria := bson.M{"tglDatang": bson.M{"$gte": formattedStartOfMonth, "$lt": formattedEndOfMonth}}
		countData, err := collection.CountDocuments(context.Background(), filterCriteria)
		if err != nil {
			somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Counting went wrong", "statusCode": 400})
			w.Write(somethingWentWrong)
			return
		}
		jsonData, _ := json.Marshal(map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": countData})
		w.Write(jsonData)
		return
	case 1:
		collection := db.Collection("soap_kehamilan")
		now := time.Now().UTC() // Declare the "now" variable
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		filterCriteria := bson.M{"soapAnc.tanggal": bson.M{"$gte": startOfMonth, "$lt": endOfMonth}}
		countData, err := collection.CountDocuments(context.Background(), filterCriteria)
		if err != nil {
			somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Counting went wrong", "statusCode": 400})
			w.Write(somethingWentWrong)
			return
		}
		jsonData, _ := json.Marshal(map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": countData})
		w.Write(jsonData)
		return

	case 2:
		collection := db.Collection("soap_imunisasi")
		now := time.Now().UTC() // Declare the "now" variable
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		filterCriteria := bson.M{"tglDatang": bson.M{"$gte": startOfMonth, "$lt": endOfMonth}}
		countData, err := collection.CountDocuments(context.Background(), filterCriteria)
		if err != nil {
			somethingWentWrong, _ := json.Marshal(map[string]interface{}{"message": "Counting went wrong", "statusCode": 400})
			w.Write(somethingWentWrong)
			return
		}
		jsonData, _ := json.Marshal(map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": countData})
		w.Write(jsonData)
		return

	default:
		jsonData, _ := json.Marshal(map[string]interface{}{"statusCode": 200, "message": "Success", "jumlah": 0})
		w.Write(jsonData)
		return
	}
}
