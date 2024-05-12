package getpasien

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetPasien(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		message := map[string]string{"message": "Error, .env URI NOT FOUND", "statusCode": "401"}
		jsonData, _ := json.Marshal(message)

		w.Write(jsonData)
		return
	}

	id_pasien := r.URL.Query().Get("id_pasien")
	id_pasien_int, err := strconv.Atoi(id_pasien)
	if err != nil {
		message := map[string]string{"message": "Invalid id_pasien", "statusCode": "400"}
		jsonData, _ := json.Marshal(message)

		w.Write(jsonData)
		return
	}

	client, err := mongo.Connect(context.Background(), options.Client().
		ApplyURI(uri))
	if err != nil {
		somethingwentwrong, _ := json.Marshal(map[string]interface{}{"message": "Something went wrong", "statusCode": 400})
		w.Write(somethingwentwrong)
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			somethingwentwrong, _ := json.Marshal(map[string]interface{}{"message": "Something went wrong", "statusCode": 400})
			w.Write(somethingwentwrong)
			panic(err)
		}
	}()

	coll := client.Database("mydb").Collection("pasien")
	var result bson.M
	err = coll.FindOne(context.Background(), bson.D{{"id_pasien", id_pasien_int}}).Decode(&result)

	if err == mongo.ErrNoDocuments {
		message := map[string]string{"message": "No user found", "statusCode": "200"}
		jsonData, _ := json.Marshal(message)

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}
	if err != nil {
		panic(err)
	}

	jsonData, err := json.Marshal(map[string]interface{}{"message": "Success", "data": result, "statusCode": "200"})
	if err != nil {
		somethingwentwrong, _ := json.Marshal(map[string]interface{}{"message": "Something went wrong", "statusCode": 400})
		w.Write(somethingwentwrong)
		panic(err)
	}
	w.Write(jsonData)
}
