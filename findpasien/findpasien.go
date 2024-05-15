package findpasien

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func PasienPerLayanan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := godotenv.Load()
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": ".env not found"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "error connecting to database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "Database disconnected"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
		}
	}()
	db := client.Database("mydb")
	collection := db.Collection("pasien")

	keyword := r.URL.Query().Get("keyword")
	id_layanan_raw := r.URL.Query().Get("id_layanan")
	id_layanan, err := strconv.Atoi(id_layanan_raw)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Id_layanan needed"})
		w.Write(jsonData)
		return
	}

	var regexquery bson.M
	if id_layanan == 0 {
		regexquery = bson.M{
			"$and": []bson.M{
				{"data_kb": bson.M{"$exists": true}},
				{"nama_pasien": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
	} else if id_layanan == 1 {
		regexquery = bson.M{
			"$and": []bson.M{
				{"data_kehamilan": bson.M{"$exists": true}},
				{"nama_pasien": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}

	} else if id_layanan == 2 {
		regexquery = bson.M{
			"$and": []bson.M{
				{"data_imunisasi": bson.M{"$exists": true}},
				{"nama_pasien": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}

	} else {
		jsonData, _ := json.Marshal(map[string]string{"message": "Under construction"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	cursor, err := collection.Find(context.Background(), regexquery)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Error executing query"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}
	defer cursor.Close(context.Background())
	finalList := []int{}
	for cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			jsonData, _ := json.Marshal(map[string]string{"message": "Error decoding results"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}
		// get id_pasien from result (id_pasien is int in the database mongodb)
		idPasien_raw, ok := result["id_pasien"].(int64)

		if !ok {
			// log.Printf("id_pasien data type: %T", result["id_pasien"])
			jsonData, _ := json.Marshal(map[string]string{"message": "Failed to convert id_pasien to int"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
			return
		}
		idPasien := int(idPasien_raw)
		finalList = append(finalList, idPasien)
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "Success", "id_pasien_list": finalList})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}