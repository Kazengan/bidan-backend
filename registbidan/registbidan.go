package registbidan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Password    string `bson:"password" json:"password"`
	FullName    string `bson:"full_name" json:"full_name"`
	Username    string `bson:"username" json:"username"`
	PhoneNumber string `bson:"phone_number" json:"phone_number"`
	Role        string `bson:"role" json:"role"`
}

func isPasswordValid(password string) (bool, error) {
	if len(password) < 8 {
		return false, fmt.Errorf("password must be at least 8 characters long")
	}

	var (
		hasUppercase bool
		hasLowercase bool
		hasDigit     bool
		hasSpecial   bool
	)

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUppercase = true
		case 'a' <= char && char <= 'z':
			hasLowercase = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case !strings.ContainsAny(string(char), "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"):
			hasSpecial = true
		}
	}
	return hasUppercase && hasLowercase && hasDigit && hasSpecial, nil
}

func RegistBidan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": ".env not found"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error connecting to database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	// Decode request body into User struct
	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "request body decode error"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	passwordValid, err := isPasswordValid(user.Password)
	if !passwordValid {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	db := client.Database("mydb")
	collection := db.Collection("bidan")

	//check in database if username already exists
	filter := bson.M{"$or": []bson.M{
		{"username": user.Username},
	}}

	var existingUser bson.M
	err = collection.FindOne(context.Background(), filter).Decode(&existingUser)
	if err == nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "username already exists"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	// Insert new user document but hash the password to bcrypt first
	// the references is this code     hashed_password = bcrypt.hashpw(password.encode("utf-8"), bcrypt.gensalt())

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error hashing password"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	user.Password = string(hashedPassword)
	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "error inserting user document"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "user registered successfully"})
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)

}
