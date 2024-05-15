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

	db := client.Database("mydb")
	collection := db.Collection("bidan")

	//filter ["role"] not "superadmin"
	filter := bson.M{"role": bson.M{"$ne": "superadmin"}}

	cursor, err := collection.Find(context.Background(), filter, options.Find().SetProjection(bson.M{"password": false}))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error fetching bidan data"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	var bidanData []bson.M
	if err = cursor.All(context.Background(), &bidanData); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error decoding bidan data"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "Success", "data": bidanData})
	w.Write(jsonData)
}
