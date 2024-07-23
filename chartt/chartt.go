package chartt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	clientOptions := options.Client().ApplyURI(os.Getenv("AZURE_MONGODB_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	return client, nil
}

func respondWithError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": message}
	if err != nil {
		response["error"] = err.Error()
	}
	jsonData, _ := json.Marshal(response)
	w.WriteHeader(status)
	w.Write(jsonData)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, _ := json.Marshal(payload)
	w.WriteHeader(status)
	w.Write(jsonData)
}

func Chartt(w http.ResponseWriter, r *http.Request) {
	client, err := connectToDatabase()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to database", err)
		return
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database disconnect error", err)
			return
		}
	}()

	db := client.Database("mydb")

	idLayananStr := r.URL.Query().Get("id_layanan")
	var idLayanan int
	if idLayananStr != "" {
		idLayanan, err = strconv.Atoi(idLayananStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid id_layanan", err)
			return
		}
	}

	currentYear := time.Now().Year()
	filterCurrentYear := strconv.Itoa(currentYear) + "-"

	var pipeline []bson.M

	if idLayananStr == "" {
		pipeline = []bson.M{
			{
				"$project": bson.M{
					"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 7}},
					"id_layanan": bson.M{"$literal": 0},
				},
			},
			{
				"$unionWith": bson.M{
					"coll": "soap_kehamilan",
					"pipeline": []bson.M{
						{
							"$project": bson.M{
								"tanggal":    bson.M{"$substr": []interface{}{"$soapAnc.tanggal", 0, 7}},
								"id_layanan": bson.M{"$literal": 1},
							},
						},
					},
				},
			},
			{
				"$unionWith": bson.M{
					"coll": "soap_imunisasi",
					"pipeline": []bson.M{
						{
							"$project": bson.M{
								"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 7}},
								"id_layanan": bson.M{"$literal": 2},
							},
						},
					},
				},
			},
		}
	} else {
		switch idLayanan {
		case 0:
			pipeline = []bson.M{
				{
					"$project": bson.M{
						"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 7}},
						"id_layanan": bson.M{"$literal": 0},
					},
				},
			}
		case 1:
			pipeline = []bson.M{
				{
					"$project": bson.M{
						"tanggal":    bson.M{"$substr": []interface{}{"$soapAnc.tanggal", 0, 7}},
						"id_layanan": bson.M{"$literal": 1},
					},
				},
			}
		case 2:
			pipeline = []bson.M{
				{
					"$project": bson.M{
						"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 7}},
						"id_layanan": bson.M{"$literal": 2},
					},
				},
			}
		default:
			respondWithError(w, http.StatusBadRequest, "Invalid id_layanan", nil)
			return
		}
	}

	pipeline = append(pipeline, []bson.M{
		{
			"$match": bson.M{
				"tanggal": bson.M{"$regex": "^" + filterCurrentYear},
			},
		},
		{
			"$project": bson.M{
				"tanggal":    1,
				"id_layanan": 1,
				"bulan":      bson.M{"$substr": []interface{}{"$tanggal", 5, 2}},
			},
		},
		{
			"$group": bson.M{
				"_id":    "$bulan",
				"jumlah": bson.M{"$sum": 1},
			},
		},
	}...)

	var collection *mongo.Collection
	switch idLayanan {
	case 0:
		collection = db.Collection("soap_kb")
	case 1:
		collection = db.Collection("soap_kehamilan")
	case 2:
		collection = db.Collection("soap_imunisasi")
	default:
		collection = db.Collection("soap_kb")
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error aggregating data", err)
		return
	}
	defer cursor.Close(context.TODO())

	resultMap := []bson.M{
		{"month": "Jan", "revenue": 0},
		{"month": "Feb", "revenue": 0},
		{"month": "Mar", "revenue": 0},
		{"month": "Apr", "revenue": 0},
		{"month": "May", "revenue": 0},
		{"month": "Jun", "revenue": 0},
		{"month": "Jul", "revenue": 0},
		{"month": "Aug", "revenue": 0},
		{"month": "Sep", "revenue": 0},
		{"month": "Oct", "revenue": 0},
		{"month": "Nov", "revenue": 0},
		{"month": "Dec", "revenue": 0},
	}

	for cursor.Next(context.Background()) {
		var entry struct {
			Month  string `bson:"_id"`
			Jumlah int    `bson:"jumlah"`
		}
		if err := cursor.Decode(&entry); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error decoding data", err)
			return
		}
		month_int, err := strconv.Atoi(entry.Month)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error converting month to int", err)
			return
		}
		resultMap[month_int-1]["revenue"] = entry.Jumlah
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"data": resultMap, "message": "Success"})
}
