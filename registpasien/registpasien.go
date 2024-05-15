package registpasien

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
}

func sendVerificationEmail(from, password, to_email, fullname, verif string) error {
	subject := "[BidanMandiri] KODE VERIFIKASI EMAIL"
	body := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>

    <meta charset="utf-8">
    <meta http-equiv="x-ua-compatible" content="ie=edge">
    <title>Email Confirmation</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style type="text/css">
    /**
    * Google webfonts. Recommended to include the .woff version for cross-client compatibility.
    */
    @media screen {{
        @font-face {{
        font-family: 'Source Sans Pro';
        font-style: normal;
        font-weight: 400;
        src: local('Source Sans Pro Regular'), local('SourceSansPro-Regular'), url(https://fonts.gstatic.com/s/sourcesanspro/v10/ODelI1aHBYDBqgeIAH2zlBM0YzuT7MdOe03otPbuUS0.woff) format('woff');
        }}
        @font-face {{
        font-family: 'Source Sans Pro';
        font-style: normal;
        font-weight: 700;
        src: local('Source Sans Pro Bold'), local('SourceSansPro-Bold'), url(https://fonts.gstatic.com/s/sourcesanspro/v10/toadOcfmlt9b38dHJxOBGFkQc6VGVFSmCnC_l7QZG60.woff) format('woff');
        }}
    }}
    /**
    * Avoid browser level font resizing.
    * 1. Windows Mobile
    * 2. iOS / OSX
    */
    body,
    table,
    td,
    a {{
        -ms-text-size-adjust: 100%%; /* 1 */
        -webkit-text-size-adjust: 100%%; /* 2 */
    }}
    /**
    * Remove extra space added to tables and cells in Outlook.
    */
    table,
    td {{
        mso-table-rspace: 0pt;
        mso-table-lspace: 0pt;
    }}
    /**
    * Better fluid images in Internet Explorer.
    */
    img {{
        -ms-interpolation-mode: bicubic;
    }}
    /**
    * Remove blue links for iOS devices.
    */
    a[x-apple-data-detectors] {{
        font-family: inherit !important;
        font-size: inherit !important;
        font-weight: inherit !important;
        line-height: inherit !important;
        color: inherit !important;
        text-decoration: none !important;
    }}
    /**
    * Fix centering issues in Android 4.4.
    */
    div[style*="margin: 16px 0;"] {{
        margin: 0 !important;
    }}
    body {{
        width: 100%% !important;
        height: 100%% !important;
        padding: 0 !important;
        margin: 0 !important;
    }}
    /**
    * Collapse table borders to avoid space between cells.
    */
    table {{
        border-collapse: collapse !important;
    }}
    a {{
        color: #1a82e2;
    }}
    img {{
        height: auto;
        line-height: 100%%;
        text-decoration: none;
        border: 0;
        outline: none;
    }}
    </style>

    </head>
    <body style="background-color: #e9ecef;">

    <!-- start preheader -->
    <div class="preheader" style="display: none; max-width: 0; max-height: 0; overflow: hidden; font-size: 1px; line-height: 1px; color: #fff; opacity: 0;">
        Hi %s, terima kasih telah mendaftar pada layanan BidanMandiri. Silakan tekan tombol verifikasi di bawah ini untuk menyelesaikan proses pendaftaran!
    </div>
    <!-- end preheader -->

    <!-- start body -->
    <table border="0" cellpadding="0" cellspacing="0" width="100%%">

        <!-- start hero -->
        <tr>
        <td align="center" bgcolor="#e9ecef">
            <!--[if (gte mso 9)|(IE)]>
            <table align="center" border="0" cellpadding="0" cellspacing="0" width="600">
            <tr>
            <td align="center" valign="top" width="600">
            <![endif]-->
            <table border="0" cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin-top: 30px;">
            <tr>
                <td align="left" bgcolor="#ffffff" style="padding: 36px 24px 0; font-family: 'Source Sans Pro', Helvetica, Arial, sans-serif; border-top: 3px solid #d4dadf;">
                <h1 style="margin: 0; font-size: 32px; font-weight: 700; letter-spacing: -1px; line-height: 48px;">Konfirmasi Email Anda</h1>
                </td>
            </tr>
            </table>
            <!--[if (gte mso 9)|(IE)]>
            </td>
            </tr>
            </table>
            <![endif]-->
        </td>
        </tr>
        <!-- end hero -->

        <!-- start copy block -->
        <tr>
        <td align="center" bgcolor="#e9ecef">
            <!--[if (gte mso 9)|(IE)]>
            <table align="center" border="0" cellpadding="0" cellspacing="0" width="600">
            <tr>
            <td align="center" valign="top" width="600">
            <![endif]-->
            <table border="0" cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px;">

            <!-- start copy -->
            <tr>
                <td align="left" bgcolor="#ffffff" style="padding: 24px; font-family: 'Source Sans Pro', Helvetica, Arial, sans-serif; font-size: 16px; line-height: 24px;">
                <p style="margin: 0;">Tekan tombol di bawah ini untuk mengonfirmasi alamat email Anda. Jika Anda tidak membuat merasa melakukan pembuatan akun pada BidanMandiri, Anda bisa mengabaikan dan menghapus email ini.</p>
                </td>
            </tr>
            <!-- end copy -->

            <!-- start button -->
            <tr>
                <td align="left" bgcolor="#ffffff">
                <table border="0" cellpadding="0" cellspacing="0" width="100%%">
                    <tr>
                    <td align="center" bgcolor="#ffffff" style="padding: 12px;">
                        <table border="0" cellpadding="0" cellspacing="0">
                        <tr>
                            <td align="center" bgcolor="#1a82e2" style="border-radius: 6px;">
                            <a href="https://faas-sgp1-18bc02ac.doserverless.co/api/v1/web/fn-0100bb5b-93c0-4b54-ae68-1b56a41e7f49/verif_endpoint/verif?email=%s&verification_code=%s" target="_blank" style="display: inline-block; padding: 16px 36px; font-family: 'Source Sans Pro', Helvetica, Arial, sans-serif; font-size: 16px; color: #ffffff; text-decoration: none; border-radius: 6px;">Verifikasi Email</a>
                            </td>
                        </tr>
                        </table>
                    </td>
                    </tr>
                </table>
                </td>
            </tr>
            <!-- end button -->

            <!-- start copy -->
            <tr>
                <td align="left" bgcolor="#ffffff" style="padding: 24px; font-family: 'Source Sans Pro', Helvetica, Arial, sans-serif; font-size: 16px; line-height: 24px;">
                <p style="margin: 0;">Jika tidak berhasil, salin dan tempel tautan berikut ini di browser anda:</p>
                <p style="margin: 0;"><a href="https://faas-sgp1-18bc02ac.doserverless.co/api/v1/web/fn-0100bb5b-93c0-4b54-ae68-1b56a41e7f49/verif_endpoint/verif?email=%s&verification_code=%s" target="_blank">https://faas-sgp1-18bc02ac.doserverless.co/api/v1/web/fn-0100bb5b-93c0-4b54-ae68-1b56a41e7f49/verif_endpoint/verif?email=%s&verification_code=%s</a></p>
                </td>
            </tr>
            <!-- end copy -->

            <!-- start copy -->
            <tr>
                <td align="left" bgcolor="#ffffff" style="padding: 24px; font-family: 'Source Sans Pro', Helvetica, Arial, sans-serif; font-size: 16px; line-height: 24px; border-bottom: 3px solid #d4dadf">
                <p style="margin: 0;">Salam hangat,<br> BidanMandiri</p>
                </td>
            </tr>
            <!-- end copy -->

            </table>
            <!--[if (gte mso 9)|(IE)]>
            </td>
            </tr>
            </table>
            <![endif]-->
        </td>
        </tr>
        <!-- end copy block -->

        <!-- start footer -->
        <tr>
        <td align="center" bgcolor="#e9ecef" style="padding: 24px;">
            <!--[if (gte mso 9)|(IE)]>
            <table align="center" border="0" cellpadding="0" cellspacing="0" width="600">
            <tr>
            <td align="center" valign="top" width="600">
            <![endif]-->
            <table border="0" cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px;">

            </table>
            <!--[if (gte mso 9)|(IE)]>
            </td>
            </tr>
            </table>
            <![endif]-->
        </td>
        </tr>
        <!-- end footer -->

    </table>
    <!-- end body -->

    </body>
    </html>
	`, fullname, to_email, verif, to_email, verif, to_email, verif)

	msg := []byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\nContent-Type: text/html; charset=UTF-8\n\n%s",
		from, to_email, subject, body))

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to_email}, msg)
	if err != nil {
		return err
	}
	return nil
}

func RegistPasien(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": ".env file not found"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Error connecting to database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			jsonData, _ := json.Marshal(map[string]interface{}{"message": "Error disconnecting from database"})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonData)
		}
	}()

	db := client.Database("mydb")
	pasien_collection := db.Collection("users")
	pending_pasien_collection := db.Collection("pending_users")

	// Decode request body into User struct
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "email, password, full name, username, and phone_number are required"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	//check in db user with email or username already exist or not
	filter := bson.M{"$or": []bson.M{
		{"email": user.Email},
		{"username": user.Username},
	}}
	//find user with filter
	count, err := pasien_collection.CountDocuments(context.Background(), filter)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Error checking user existence"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	if count > 0 {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Email or username already exists"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonData)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Error hashing password"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	//generate random number between 000000 to 999999
	verification_code := fmt.Sprintf("%06d", rand.Intn(1000000))
	// Send verification email
	from := os.Getenv("EMAIL")
	if from == "" {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "no EMAIL in .env"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	password := os.Getenv("EMAIL_PASSWORD")
	if password == "" {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "no EMAIL_PASSWORD in .env"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	//try send email to user
	err = sendVerificationEmail(from, password, user.Email, user.FullName, verification_code)
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Error sending verification email"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	// Insert pending user into MongoDB
	_, err = pending_pasien_collection.InsertOne(context.Background(), bson.M{
		"email":             user.Email,
		"password":          string(hashedPassword),
		"full_name":         user.FullName,
		"username":          user.Username,
		"phone_number":      user.PhoneNumber,
		"verification_code": verification_code,
	})
	if err != nil {
		jsonData, _ := json.Marshal(map[string]interface{}{"message": "Error inserting user into database"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	jsonData, _ := json.Marshal(map[string]interface{}{"message": "User registered successfully", "verification_code": verification_code})
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
