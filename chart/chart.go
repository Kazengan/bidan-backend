package chart

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Result struct {
	Month   string `bson:"month" json:"month"`
	Revenue int    `bson:"revenue" json:"revenue"`
}

func connectToDatabase() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
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

	pipeline := mongo.Pipeline{
		{
			{Key: "$unionWith", Value: bson.D{
				{Key: "coll", Value: "soap_kehamilan"},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$project", Value: bson.D{
						{Key: "date", Value: "$soapAnc.tanggal"},
					}}},
				}},
			}},
		},
		{
			{Key: "$unionWith", Value: bson.D{
				{Key: "coll", Value: "soap_imunisasi"},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$project", Value: bson.D{
						{Key: "date", Value: "$tglDatang"},
					}}},
				}},
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "yearMonth", Value: bson.D{
					{Key: "$dateToString", Value: bson.D{
						{Key: "format", Value: "%Y-%m"},
						{Key: "date", Value: bson.D{{Key: "$dateFromString", Value: bson.D{{Key: "dateString", Value: "$date"}}}}},
					}},
				}},
			}},
		},
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$yearMonth"},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "month", Value: bson.D{
					{Key: "$substr", Value: bson.A{"$_id", 5, 2}},
				}},
				{Key: "revenue", Value: "$count"},
			}},
		},
		{
			{Key: "$sort", Value: bson.D{{Key: "month", Value: 1}}},
		},
	}

	cursor, err := db.Collection("soap_kb").Aggregate(context.TODO(), pipeline)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error aggregating data")
		return
	}
	defer cursor.Close(context.TODO())

	var results []Result
	for cursor.Next(context.TODO()) {
		var result Result
		if err := cursor.Decode(&result); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error decoding data")
			return
		}
		results = append(results, result)
	}
	if err := cursor.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error iterating cursor")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Success",
		"data":    results,
	})
}
