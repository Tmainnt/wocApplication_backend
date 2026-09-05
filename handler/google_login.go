package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	auth "woc/auth"
	service "woc/service"
)

type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
}

func GoogleLoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service.HttpMethodPost(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req GoogleLoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid json", http.StatusBadGateway)
			log.Println(err)
			return
		}

		// TODO: Verify the ID Token with Google API here

		// Assume email is extracted from verified token
		email := "user@example.com"
		name := "Google User"

		var user User
		err = db.QueryRow(
			`SELECT user_id, user_name, user_email, user_role 
	 FROM users WHERE user_email=$1`,
			email,
		).Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Role,
		)

		if err == sql.ErrNoRows {
			// Register user if not exists
			err = db.QueryRow(
				"INSERT INTO users(user_name, user_email, user_role, user_status) VALUES($1, $2, $3, $4) RETURNING user_id",
				name,
				email,
				"user",
				"active",
			).Scan(&user.ID)
			
			if err != nil {
				http.Error(w, "Database Error", 500)
				log.Println(err)
				return
			}
		} else if err != nil {
			http.Error(w, "Database Error", 500)
			log.Println(err)
			return
		}

		// Generate tokens
		token, _ := auth.GenerateAccessToken(user.ID, user.Email, user.Role)
		refreshToken, _ := auth.GenerateRefreshToken(user.ID, user.Email, user.Role)

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
