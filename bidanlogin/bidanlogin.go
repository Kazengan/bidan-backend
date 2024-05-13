package bidanlogin

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		return
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		message := map[string]string{"message": "Error, .env URI NOT FOUNDD"}
		jsonData, _ := json.Marshal(message)

		w.WriteHeader(401)
		w.Write(jsonData)
		return
	}

	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	// Check if the required fields are provided
	if username == "" || password == "" {
		message := map[string]string{"message": "Username and password are required"}
		jsonData, _ := json.Marshal(message)

		w.WriteHeader(400)
		w.Write(jsonData)
		return
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		somethingwentwrong, _ := json.Marshal(map[string]interface{}{"message": "Something went wrong"})

		w.WriteHeader(400)
		w.Write(somethingwentwrong)
		return
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			somethingwentwrong, _ := json.Marshal(map[string]interface{}{"message": "Something went wrong"})

			w.WriteHeader(400)
			w.Write(somethingwentwrong)
			return
		}
	}()

	users_collection := client.Database("mydb").Collection("bidan")

	var user bson.M
	err = users_collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		somethingwentwrong, _ := json.Marshal(map[string]interface{}{"message": "Something went wrong"})

		w.WriteHeader(500)
		w.Write(somethingwentwrong)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user["password"].(string)), []byte(password)) != nil {
		message := map[string]string{"message": "Username or password is wrong"}

		w.WriteHeader(401)
		jsonData, _ := json.Marshal(message)

		w.Write(jsonData)
		return
	}

	delete(user, "password")
	jsonData, err := json.Marshal(map[string]interface{}{"message": "Login successful", "data": user, "statusCode": 200})
	if err != nil {
		somethingwentwrong, _ := json.Marshal(map[string]interface{}{"message": "Something went wrong"})
		w.WriteHeader(500)
		w.Write(somethingwentwrong)
		return
	}
	w.Write(jsonData)
}
