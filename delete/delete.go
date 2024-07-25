package delete

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

func Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Connect database error"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}
	defer client.Disconnect(context.Background())

	idPasienStr := r.URL.Query().Get("id_pasien")
	idPasienInt, err := strconv.Atoi(idPasienStr)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Invalid id_pasien"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	db := client.Database("mydb")
	pasien_collection := db.Collection("pasien")
	kb_collection := db.Collection("soap_kb")
	kehamilan_collection := db.Collection("soap_kehamilan")
	imunisasi_collection := db.Collection("soap_imunisasi")

	filter := bson.M{"id_pasien": idPasienInt}

	session, err := client.StartSession()
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Error starting session"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}
	defer session.EndSession(context.Background())

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, err := pasien_collection.DeleteOne(sessCtx, filter)
		if err != nil {
			return nil, err
		}

		_, err = kb_collection.DeleteOne(sessCtx, filter)
		if err != nil {
			return nil, err
		}

		_, err = kehamilan_collection.DeleteOne(sessCtx, filter)
		if err != nil {
			return nil, err
		}

		_, err = imunisasi_collection.DeleteOne(sessCtx, filter)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	_, err = session.WithTransaction(context.Background(), callback)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]string{"message": "Transaction error"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]string{"message": "Delete successful"})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
