package chart

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectToDatabase() (*mongo.Client, error) {

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
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

func Chart(w http.ResponseWriter, r *http.Request) {
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
	}

	currentYear := time.Now().Year()
	filterCurrentYear := strconv.Itoa(currentYear) + "-"

	var pipeline []bson.M

	if idLayananStr == "" {
		// If id_layanan is not provided, aggregate from all collections
		pipeline = []bson.M{
			{
				"$project": bson.M{
					"_id":        0,
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
								"_id":        0,
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
								"_id":        0,
								"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 7}},
								"id_layanan": bson.M{"$literal": 2},
							},
						},
					},
				},
			},
		}
	} else {
		// If id_layanan is provided, aggregate from the specific collection
		switch idLayanan {
		case 0:
			pipeline = []bson.M{
				{
					"$project": bson.M{
						"_id":        0,
						"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 7}},
						"id_layanan": bson.M{"$literal": 0},
					},
				},
			}
		case 1:
			pipeline = []bson.M{
				{
					"$project": bson.M{
						"_id":        0,
						"tanggal":    bson.M{"$substr": []interface{}{"$soapAnc.tanggal", 0, 7}},
						"id_layanan": bson.M{"$literal": 1},
					},
				},
			}
		case 2:
			pipeline = []bson.M{
				{
					"$project": bson.M{
						"_id":        0,
						"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 7}},
						"id_layanan": bson.M{"$literal": 2},
					},
				},
			}
		default:
			respondWithError(w, http.StatusBadRequest, "Invalid id_layanan")
			return
		}
	}

	// Append the common stages for all cases
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

	// Determine the collection to use based on id_layanan
	var collection *mongo.Collection
	switch idLayanan {
	case 0:
		collection = db.Collection("soap_kb")
	case 1:
		collection = db.Collection("soap_kehamilan")
	case 2:
		collection = db.Collection("soap_imunisasi")
	default:
		collection = db.Collection("soap_kb") // Default to soap_kb if id_layanan is not provided
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error aggregating data")
		return
	}
	defer cursor.Close(context.TODO())

	// Initialize the result map with default values for all months
	resultMap := []bson.M{
		{"month": "Janw", "revenue": 0},
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

	// Iterate through the cursor and update the resultMap
	for cursor.Next(context.Background()) {
		var entry struct {
			Month  string `bson:"_id"`
			Jumlah int    `bson:"jumlah"`
		}
		if err := cursor.Decode(&entry); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error decoding data")
			return
		}
		month_int, err := strconv.Atoi(entry.Month)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error converting month to int")
			return
		}
		resultMap[month_int-1]["revenue"] = entry.Jumlah
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"data": resultMap, "message": "Success"})
}
