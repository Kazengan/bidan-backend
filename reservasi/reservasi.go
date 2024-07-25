package reservasi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// calculateReminderTime parses the reservation date and returns the timestamp
func calculateReminderTime(hariReservasi string) (int64, error) {
	layout := "2006-01-02"
	t, err := time.Parse(layout, hariReservasi)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// Reservasi handles the reservation request
func Reservasi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var args map[string]string
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nama := args["nama"]
	phoneNumber := args["noHP"]
	idLayanan := args["id_layanan"]
	hariReservasi := args["hariReservasi"][:10]
	waktu := args["waktuTersedia"]

	if nama == "" {
		http.Error(w, `{"message": "nama needed"}`, http.StatusUnauthorized)
		return
	}
	if phoneNumber == "" {
		http.Error(w, `{"message": "phone number needed"}`, http.StatusUnauthorized)
		return
	}
	if idLayanan == "" {
		http.Error(w, `{"message": "id_layanan needed"}`, http.StatusUnauthorized)
		return
	}
	if hariReservasi == "" {
		http.Error(w, `{"message": "hari reservasi needed"}`, http.StatusUnauthorized)
		return
	}
	if waktu == "" {
		http.Error(w, `{"message": "waktu tersedia needed"}`, http.StatusUnauthorized)
		return
	}

	idLayananInt, err := strconv.Atoi(idLayanan)
	if err != nil {
		http.Error(w, `{"message": "invalid id_layanan"}`, http.StatusUnauthorized)
		return
	}

	mongodbURI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(mongodbURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "error %v"}`, err), http.StatusUnauthorized)
		return
	}
	defer client.Disconnect(context.Background())

	db := client.Database("mydb")
	reservasiCollection := db.Collection("reservasi_layanan")
	reminderCollection := db.Collection("reminder")

	remindTimestamp, err := calculateReminderTime(hariReservasi)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "error %v"}`, err), http.StatusUnauthorized)
		return
	}

	jsonData1 := bson.M{
		"nama":          nama,
		"noHP":          phoneNumber,
		"id_layanan":    idLayananInt,
		"hariReservasi": hariReservasi,
		"waktuTersedia": waktu,
	}

	jsonData2 := bson.M{
		"nama":             nama,
		"noHP":             phoneNumber,
		"id_layanan":       idLayananInt,
		"remind_timestamp": remindTimestamp,
		"status":           "reminder reservasi",
	}

	session, err := client.StartSession()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "error %v"}`, err), http.StatusUnauthorized)
		return
	}
	defer session.EndSession(context.Background())

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, err := reservasiCollection.InsertOne(sessCtx, jsonData1)
		if err != nil {
			return nil, err
		}
		_, err = reminderCollection.InsertOne(sessCtx, jsonData2)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	_, err = session.WithTransaction(context.Background(), callback)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "transaction failed: %v"}`, err), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "success"}`))
}
