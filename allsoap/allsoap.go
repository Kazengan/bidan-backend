package allsoap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		return nil, fmt.Errorf("error connecting to database")
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

func convertToIndonesianDate(dateStr string) (string, error) {
	tanggalDatetime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", err
	}

	namaHariID := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	namaBulanID := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	hari := namaHariID[tanggalDatetime.Weekday()]
	bulan := namaBulanID[int(tanggalDatetime.Month())-1]
	tahun := tanggalDatetime.Year()

	return fmt.Sprintf("%s, %d %s %d", hari, tanggalDatetime.Day(), bulan, tahun), nil
}

func Allsoap(w http.ResponseWriter, r *http.Request) {
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

	pipeline := []bson.M{
		{
			"$project": bson.M{
				"id_pasien":  1,
				"datetime":   "$tglDatang",
				"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 10}},
				"id_layanan": bson.M{"$literal": "KB"},
			},
		},
		{
			"$unionWith": bson.M{
				"coll": "soap_kehamilan",
				"pipeline": []bson.M{
					{
						"$project": bson.M{
							"id_pasien":  1,
							"datetime":   "$soapAnc.tanggal",
							"tanggal":    bson.M{"$substr": []interface{}{"$soapAnc.tanggal", 0, 10}},
							"id_layanan": bson.M{"$literal": "Kehamilan"},
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
							"id_pasien":  1,
							"datetime":   "$tglDatang",
							"tanggal":    bson.M{"$substr": []interface{}{"$tglDatang", 0, 10}},
							"id_layanan": bson.M{"$literal": "Imunisasi"},
						},
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "pasien",
				"localField":   "id_pasien",
				"foreignField": "id_pasien",
				"as":           "pasien_info",
			},
		},
		{
			"$unwind": "$pasien_info",
		},
		{
			"$project": bson.M{
				"id_pasien":  1,
				"tanggal":    1,
				"id_layanan": 1,
				"datetime":   1,
				"namaPasien": "$pasien_info.nama_pasien",
				"noHP":       "$pasien_info.no_hp",
			},
		},
		{
			"$sort": bson.M{
				"tanggal": 1, // 1 for ascending order, -1 for descending order
			},
		},
	}

	collection := db.Collection("soap_kb")
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error aggregating data", err)
		return
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading cursor data", err)
		return
	}

	for i, result := range results {
		if tanggal, ok := result["tanggal"].(string); ok {
			indonesianDate, err := convertToIndonesianDate(tanggal)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error converting date", err)
				return
			}
			results[i]["tanggal"] = indonesianDate
		}
	}

	respondWithJSON(w, http.StatusOK, results)
}
