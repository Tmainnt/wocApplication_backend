package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	service "woc/service"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Name     string `json:"user_name"`
	Password string `json:"user_pass"`
	Email    string `json:"user_email"`
	Gender   string `json:"user_gender"`
	DOF      string `json:"date_of_birth"`
	PhoneNB  string `json:"user_phone"`
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service.HttpMethodPost(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid json", http.StatusBadGateway)
			log.Println("JSON Decode Error:", err)
			return
		}

		if req.Name == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "Email and password required", 400)
			log.Println(err)
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Hash error", 500)
			log.Println(err)
			return
		}

		_, err = db.Exec(
			"INSERT INTO users(user_name, user_pass, user_email, user_gender, date_of_birth, user_phone, user_role, user_profile_image, user_background_image ,user_status) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
			req.Name,
			string(hashed),
			req.Email,
			req.Gender,
			req.DOF,
			req.PhoneNB,
			"user",
			"",
			"",
			"offline",
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("register success"))
	}
}
