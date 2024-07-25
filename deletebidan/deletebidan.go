package deletebidan

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DeleteBidan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error connecting to database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	db := client.Database("mydb")
	collection := db.Collection("bidan")

	idBidan := r.URL.Query().Get("id_bidan")
	if idBidan == "" {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "id parameter is required"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	//filter idbidan with type of objectid
	objID, err := primitive.ObjectIDFromHex(idBidan)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "invalid id format"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}
	filter := bson.M{"_id": objID}

	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error deleting bidan"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	if result.DeletedCount == 0 {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "bidan not found"})
		w.WriteHeader(http.StatusNotFound)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "bidan deleted successfully"})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
