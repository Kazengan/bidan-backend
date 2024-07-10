package export

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RequestBody struct {
	IdLayanan int               `json:"id_layanan"`
	Date      map[string]string `json:"date"`
}

func Export(w http.ResponseWriter, r *http.Request) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		http.Error(w, "Error loading .env file", http.StatusInternalServerError)
		return
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		http.Error(w, fmt.Sprintf("MongoDB connection error: %s", err.Error()), http.StatusUnauthorized)
		return
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mydb")

	// Parse request body
	var reqBody RequestBody
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	idLayanan := reqBody.IdLayanan
	dateRange := reqBody.Date

	if dateRange == nil {
		http.Error(w, "date not provided", http.StatusUnauthorized)
		return
	}

	dateFrom := dateRange["from"]
	dateTo := dateRange["to"]

	if dateFrom == "" || dateTo == "" {
		http.Error(w, "invalid date range", http.StatusUnauthorized)
		return
	}

	// Build the query
	query := bson.M{
		"tanggal_register": bson.M{
			"$gte": dateFrom,
			"$lte": dateTo,
		},
	}

	switch idLayanan {
	case 0:
		query["data_kb"] = bson.M{"$exists": true}
	case 1:
		query["data_kehamilan"] = bson.M{"$exists": true}
	case 2:
		query["data_imunisasi"] = bson.M{"$exists": true}
	default:
		http.Error(w, "id_layanan not supported", http.StatusUnauthorized)
		return
	}

	// Execute the query
	collection := db.Collection("pasien")
	cursor, err := collection.Find(context.TODO(), query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var documents []bson.M
	if err = cursor.All(context.TODO(), &documents); err != nil {
		http.Error(w, fmt.Sprintf("Cursor error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	for i := range documents {
		documents[i]["_id"] = documents[i]["_id"].(primitive.ObjectID).Hex()
	}

	// Process and format the data based on id_layanan
	var formattedDocuments []bson.M
	for _, doc := range documents {
		var returnData bson.M
		switch idLayanan {
		case 0:
			returnData = bson.M{
				"generalInformation": bson.M{
					"noFaskes":          doc["data_kb"].(bson.M)["no_faskes"],
					"noSeriKartu":       doc["data_kb"].(bson.M)["no_seri_kartu"],
					"tglDatang":         doc["tanggal_register"],
					"namaPeserta":       doc["nama_pasien"],
					"tglLahir":          doc["tanggal_lahir"],
					"usia":              doc["umur"],
					"namaPasangan":      doc["nama_pasangan"],
					"jenisPasangan":     doc["jenis_pasangan"],
					"pendidikanAkhir":   doc["pendidikan"],
					"alamat":            doc["alamat"],
					"pekerjaanPasangan": doc["pekerjaan_pasangan"],
					"statusJkn":         doc["data_kb"].(bson.M)["status_jkn"],
					"noHP":              doc["no_hp"],
				},
				"otherInformation": doc["data_kb"].(bson.M)["informasi_lainnya"],
				"skrining":         doc["data_kb"].(bson.M)["skrining"],
				"hasil":            doc["data_kb"].(bson.M)["hasil"],
				"penapisanKB":      doc["data_kb"].(bson.M)["penapisan_kb"],
			}
		case 1:
			returnData = bson.M{
				"generalInformation": bson.M{
					"agama":           doc["data_kehamilan"].(bson.M)["agama"],
					"pekerjaan":       doc["data_kehamilan"].(bson.M)["pekerjaan"],
					"desa":            doc["data_kehamilan"].(bson.M)["desa"],
					"kabupaten":       doc["data_kehamilan"].(bson.M)["kabupaten"],
					"kecamatan":       doc["data_kehamilan"].(bson.M)["kecamatan"],
					"provinsi":        doc["data_kehamilan"].(bson.M)["provinsi"],
					"rtrw":            doc["data_kehamilan"].(bson.M)["rtrw"],
					"noIbu":           doc["data_kehamilan"].(bson.M)["no_ibu"],
					"tanggalRegister": doc["tanggal_register"],
					"namaLengkap":     doc["nama_pasien"],
					"tanggalLahir":    doc["tanggal_lahir"],
					"umur":            doc["umur"],
					"namaSuami":       doc["nama_pasangan"],
					"pendidikan":      doc["pendidikan"],
					"alamatDomisili":  doc["alamat"],
				},
				"kunjunganNifas":                        doc["data_kehamilan"].(bson.M)["kunjungan_nifas"],
				"mendeteksiFaktorResikoDanResikoTinggi": doc["data_kehamilan"].(bson.M)["faktor_resiko_resiko_tinggi"],
				"pemeriksaanPNC":                        doc["data_kehamilan"].(bson.M)["pemeriksaan_pnc"],
				"persalinan":                            doc["data_kehamilan"].(bson.M)["persalinan"],
				"rencanaPersalinan":                     doc["data_kehamilan"].(bson.M)["rencana_persalinan"],
				"riwayatKehamilan":                      doc["data_kehamilan"].(bson.M)["riwayat_kehamilan"],
				"skriningTT":                            doc["data_kehamilan"].(bson.M)["skrining_tt"],
				"section2":                              doc["data_kehamilan"].(bson.M)["section2"],
			}
		case 2:
			returnData = bson.M{
				"generalInformation": bson.M{
					"nomorBayi": doc["nomor_bayi"],
					"nomor":     doc["nomor"],
					"namaBayi":  doc["nama_pasien"],
					"namaAyah":  doc["nama_ayah"],
					"usiaAyah":  doc["umur_ayah"],
					"namaIbu":   doc["nama_ibu"],
					"usiaIbu":   doc["umur_ibu"],
					"puskesmas": doc["puskesmas"],
					"bidan":     doc["bidan"],
					"alamat":    doc["alamat"],
					"desa":      doc["desa"],
					"kecamatan": doc["kecamatan"],
					"kabupaten": doc["kabupaten"],
					"provinsi":  doc["provinsi"],
					"noHP":      doc["no_hp"],
				},
				"detailBayi":                  doc["data_imunisasi"].(bson.M)["detail_bayi"],
				"pemeriksaanNeonatus":         doc["data_imunisasi"].(bson.M)["pemeriksaan_neonatus"],
				"pemeriksaanNeonatusLanjutan": doc["data_imunisasi"].(bson.M)["pemeriksaan_neonatus_lanjutan"],
				"pemeriksaanBalita":           doc["data_imunisasi"].(bson.M)["pemeriksaan_balita"],
			}
		}
		formattedDocuments = append(formattedDocuments, returnData)
	}

	// Return the formatted documents as a JSON response
	response, err := json.Marshal(map[string]interface{}{"id_layanan": idLayanan, "date": dateRange, "data": formattedDocuments})
	if err != nil {
		http.Error(w, fmt.Sprintf("JSON Marshal error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
