package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"
	auth "woc/database/auth"
	service "woc/database/service"

	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Req_Email    string `json:"user_email"`
	Req_Password string `json:"user_pass"`
}

type User struct {
	ID              int       `json:"user_id"`
	Name            string    `json:"user_name"`
	Password        string    `json:"-"`
	Email           string    `json:"user_email"`
	FName           string    `json:"first_name"`
	LName           string    `json:"last_name"`
	Gender          string    `json:"user_gender"`
	DOF             string    `json:"date_of_birth"`
	PhoneNB         string    `json:"user_phone"`
	Role            string    `json:"user_role"`
	ProfileImage    string    `json:"user_user_profile_image"`
	BackgroundImage string    `json:"user_background_image"`
	Status          string    `json:"user_status"`
	CreateTimestamp time.Time `json:"create_timestamp"`
	UpdateTimestamp time.Time `json:"update_timestamp"`
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if service.HttpMethodPost(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req loginRequest
		var user User
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid json", http.StatusBadGateway)
			log.Println(err)
			return
		}

		if req.Req_Email == "" || req.Req_Password == "" {
			http.Error(w, "Email and password required", 400)
			log.Println(err)
			return
		}

		err = db.QueryRow(
			`SELECT user_id, user_name, user_email, user_gender, date_of_birth, user_phone, user_pass, user_role, user_profile_image, user_background_image, user_status, create_timestamp, update_timestamp 
	 FROM users WHERE user_email=$1`,
			req.Req_Email,
		).Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Gender,
			&user.DOF,
			&user.PhoneNB,
			&user.Password,
			&user.Role,
			&user.ProfileImage,
			&user.BackgroundImage,
			&user.Status,
			&user.CreateTimestamp,
			&user.UpdateTimestamp,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found	", 400)
				log.Println(err)
				return
			}
			http.Error(w, "Database Error", 500)
			log.Println(err)
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(req.Req_Password))

		if err != nil {
			http.Error(w, "Invalid Email or Password, Please try again", 401)
			log.Println(err)
			return
		}

		_, err = db.Exec(
			`DELETE FROM refresh_token WHERE user_id_fk = $1`,
			user.ID,
		)
		if err != nil {
			log.Println(err)
			http.Error(w, "Can't delete old refresh token.", http.StatusInternalServerError)
			return
		}

		token, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role)
		if err != nil {
			log.Println(err)
			http.Error(w, "Can't generate token.", 500)
			return
		}

		_, err = db.Exec("UPDATE users SET user_status=$2 WHERE user_email = $1", user.Email, "active")
		if err != nil {
			log.Println(err)
			http.Error(w, "Can't update database.", 500)
			return
		}

		refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email, user.Role)
		if err != nil {
			log.Println(err)
			http.Error(w, "Can't generate token.", 500)
			return
		}

		hashToken := HashToken(refreshToken)

		_, err = db.Exec(`INSERT INTO refresh_token (user_id_fk, token, expires_timestamp)
    VALUES ($1, $2, $3)`, user.ID, hashToken, time.Now().Add(7*24*time.Hour))
		if err != nil {
			log.Println(err)
			http.Error(w, "Can't save token", 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "success",
			"access_token":  token,
			"refresh_token": refreshToken,
			"user":          user,
		})
	}
}
