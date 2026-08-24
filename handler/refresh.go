package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	auth "woc/database/auth"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func RefreshHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req RefreshRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid json", http.StatusBadRequest)
			return
		}

		// Parse JWT
		claims, err := auth.ParseRefreshToken(req.RefreshToken)
		if err != nil {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		if claims.TokenType != "RefreshToken" {
			http.Error(w, "Invalid token type", http.StatusUnauthorized)
			return
		}

		// Hash token
		hashToken := HashToken(req.RefreshToken)

		// ตรวจสอบใน DB
		var tokenInDB string
		err = db.QueryRow(`
			SELECT token
			FROM refresh_token
			WHERE user_id_fk = $1
			AND expires_timestamp > NOW()
		`, claims.UserID).Scan(&tokenInDB)

		if err != nil {
			http.Error(w, "Refresh token not found", http.StatusUnauthorized)
			return
		}

		// เปรียบเทียบ Hash
		if hashToken != tokenInDB {
			http.Error(w, "Token mismatch", http.StatusUnauthorized)
			return
		}

		// อ่าน role
		var role string

		err = db.QueryRow(`
			SELECT user_role
			FROM users
			WHERE user_id = $1
		`, claims.UserID).Scan(&role)

		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// สร้าง Token ใหม่
		newAccessToken, err := auth.GenerateAccessToken(
			claims.UserID,
			claims.Email,
			role,
		)

		if err != nil {
			http.Error(w, "Cannot generate access token", 500)
			return
		}

		newRefreshToken, err := auth.GenerateRefreshToken(
			claims.UserID,
			claims.Email,
			role,
		)

		if err != nil {
			http.Error(w, "Cannot generate refresh token", 500)
			return
		}

		// Rotate Refresh Token

		newHash := HashToken(newRefreshToken)

		_, err = db.Exec(`
			UPDATE refresh_token
			SET token = $1,
			    expires_timestamp = $2
			WHERE user_id_fk = $3
		`,
			newHash,
			time.Now().Add(7*24*time.Hour),
			claims.UserID,
		)

		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}

		// ส่งกลับ

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]string{
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken,
		})
	}
}
