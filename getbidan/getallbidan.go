package getallbidan

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAllBidan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := godotenv.Load()
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": ".env not found"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error connecting to database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}
	defer client.Disconnect(context.Background())

	db := client.Database("mydb")
	collection := db.Collection("bidan")

	// Extract the keyword from query parameters
	keyword := r.URL.Query().Get("keyword")

	// Filter to exclude "superadmin" role and match the keyword in the "name" field
	filter := bson.M{
		"role": bson.M{"$ne": "superadmin"},
	}
	if keyword != "" {
		filter["full_name"] = bson.M{"$regex": keyword, "$options": "i"}
	}

	cursor, err := collection.Find(context.Background(), filter, options.Find().SetProjection(bson.M{"password": false}))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error fetching bidan data"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}
	defer cursor.Close(context.Background())

	var bidanData []bson.M
	if err = cursor.All(context.Background(), &bidanData); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error decoding bidan data"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "Success", "data": bidanData})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
