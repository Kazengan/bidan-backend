package findpasien

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	jsonData, _ := json.Marshal(map[string]string{"message": message})
	w.WriteHeader(statusCode)
	w.Write(jsonData)
}

func buildRegexQuery(keyword string, id_layanan int) bson.M {
	var existsField string
	switch id_layanan {
	case 0:
		existsField = "data_kb"
	case 1:
		existsField = "data_kehamilan"
	case 2:
		existsField = "data_imunisasi"
	default:
		return nil
	}

	return bson.M{
		"$and": []bson.M{
			{existsField: bson.M{"$exists": true}},
			{"nama_pasien": bson.M{"$regex": keyword, "$options": "i"}},
		},
	}
}

func PasienPerLayanan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error connecting to database")
		return
	}
	defer client.Disconnect(context.Background())

	db := client.Database("mydb")
	collection := db.Collection("pasien")

	keyword := r.URL.Query().Get("keyword")
	id_layanan_raw := r.URL.Query().Get("id_layanan")
	id_layanan, err := strconv.Atoi(id_layanan_raw)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Id_layanan needed")
		return
	}

	regexquery := buildRegexQuery(keyword, id_layanan)
	if regexquery == nil {
		respondWithError(w, http.StatusInternalServerError, "Under construction")
		return
	}

	findOptions := options.Find().SetSort(bson.M{"id_pasien": -1})

	cursor, err := collection.Find(context.Background(), regexquery, findOptions)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error executing query")
		return
	}
	defer cursor.Close(context.Background())

	finalList := []int{}
	for cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error decoding results")
			return
		}

		if idPasien, ok := result["id_pasien"].(int64); ok {
			finalList = append(finalList, int(idPasien))
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to convert id_pasien to int")
			return
		}
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "Success", "id_pasien": finalList})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
