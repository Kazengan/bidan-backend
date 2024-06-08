package table

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, _ := json.Marshal(map[string]string{"message": message})
	w.WriteHeader(status)
	w.Write(jsonData)
}

func processHistoryData(pasienHistory []bson.M) {
	for _, data := range pasienHistory {
		tglDatang := data["tglDatang"].(string)
		tglDatang = tglDatang[:10]
		// convert tglDatang to dd-mm-yyyy
		tglDatang = tglDatang[8:] + "-" + tglDatang[5:7] + "-" + tglDatang[:4]
		data["tglDatang"] = tglDatang
	}
}

func convertToIndonesianDate(dateStr string) (string, error) {
	tanggalDatetime, err := time.Parse("02-01-2006", dateStr)
	if err != nil {
		return "", err
	}

	namaHariID := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	namaBulanID := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	hari := namaHariID[tanggalDatetime.Weekday()]
	bulan := namaBulanID[tanggalDatetime.Month()-1]
	tahun := tanggalDatetime.Year()

	return fmt.Sprintf("%s, %d %s %d", hari, tanggalDatetime.Day(), bulan, tahun), nil
}

func getPatientData(client *mongo.Client, idPasienArr []string, idLayananInt int) ([]bson.M, error) {
	db := client.Database("mydb")
	pasienCollection := db.Collection("pasien")

	var soapCollection *mongo.Collection
	switch idLayananInt {
	case 0:
		soapCollection = db.Collection("soap_kb")
	case 1:
		soapCollection = db.Collection("soap_kehamilan")
	case 2:
		soapCollection = db.Collection("soap_imunisasi")
	default:
		return nil, fmt.Errorf("layanan belum tersedia")
	}

	var returnData []bson.M
	for _, id := range idPasienArr {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error converting id_pasien to int64: %v", err)
		}

		pasienFilter := bson.M{"id_pasien": idInt}
		pasienInfo := pasienCollection.FindOne(context.Background(), pasienFilter)
		if pasienInfo.Err() != nil {
			return nil, fmt.Errorf("error finding pasien: %v", pasienInfo.Err())
		}

		var pasienData bson.M
		if err := pasienInfo.Decode(&pasienData); err != nil {
			return nil, fmt.Errorf("error decoding pasien_info: %v", err)
		}

		pasienHistory, err := soapCollection.Find(context.Background(), pasienFilter)
		if err != nil {
			return nil, fmt.Errorf("error finding pasien history: %v", err)
		}

		var pasienHistoryArr []bson.M
		if err = pasienHistory.All(context.Background(), &pasienHistoryArr); err != nil {
			return nil, fmt.Errorf("error decoding pasien_history: %v", err)
		}

		// Ensure subRows is an empty array if pasienHistoryArr is empty
		subRows := make([]bson.M, len(pasienHistoryArr))
		copy(subRows, pasienHistoryArr)

		if len(pasienHistoryArr) > 0 {
			processHistoryData(pasienHistoryArr)

			lastDatang := pasienHistoryArr[len(pasienHistoryArr)-1]["tglDatang"].(string)
			tanggalIndonesia, err := convertToIndonesianDate(lastDatang)
			if err != nil {
				return nil, fmt.Errorf("error converting date: %v", err)
			}

			data := bson.M{
				"id_pasien": idInt,
				"usia":      pasienData["umur"],
				"name":      pasienData["nama_pasien"],
				"tglDatang": tanggalIndonesia,
				"subRows":   subRows,
			}

			if idLayananInt == 0 {
				if dataKb, ok := pasienData["data_kb"].(bson.M); ok {
					if infoLainnya, ok := dataKb["informasi_lainnya"].(bson.M); ok {
						data["metodeKontrasepsi"] = infoLainnya["caraKBTerakhir"].(string)
					}
				}
			} else if idLayananInt == 1 {
				data["namaSuami"] = pasienData["nama_pasangan"]

			} else if idLayananInt == 2 {
				data["namaAyah"] = pasienData["nama_ayah"]
				data["namaIbu"] = pasienData["nama_ibu"]
			}

			returnData = append(returnData, data)

		} else {
			data := bson.M{
				"id_pasien": idInt,
				"name":      pasienData["nama_pasien"],
				"usia":      pasienData["umur"],
				"tglDatang": "",
				"subRows":   subRows,
			}

			if idLayananInt == 0 {
				if dataKb, ok := pasienData["data_kb"].(bson.M); ok {
					if infoLainnya, ok := dataKb["informasi_lainnya"].(bson.M); ok {
						data["metodeKontrasepsi"] = infoLainnya["caraKBTerakhir"].(string)
					}
				}
			} else if idLayananInt == 1 {
				data["namaSuami"] = pasienData["nama_pasangan"]

			} else if idLayananInt == 2 {
				data["namaAyah"] = pasienData["nama_ayah"]
				data["namaIbu"] = pasienData["nama_ibu"]
			}

			returnData = append(returnData, data)
		}
	}

	return returnData, nil
}

func Table(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	client, err := connectToDatabase()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database disconnect error")
		}
	}()

	idPasienStr := r.URL.Query().Get("id_pasien")
	if idPasienStr == "" {
		respondWithError(w, http.StatusBadRequest, "invalid id_pasien")
		return
	}

	idPasienStr = idPasienStr[1 : len(idPasienStr)-1]
	idPasienArr := strings.Split(idPasienStr, ",")

	idLayananStr := r.URL.Query().Get("id_layanan")
	if idLayananStr == "" {
		respondWithError(w, http.StatusBadRequest, "invalid id_layanan")
		return
	}

	idLayananInt, err := strconv.Atoi(idLayananStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error converting id_layanan to int")
		return
	}

	returnData, err := getPatientData(client, idPasienArr, idLayananInt)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "success", "data": returnData})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
