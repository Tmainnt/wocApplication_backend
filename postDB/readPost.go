package pdb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	auth "woc/database/auth"
)

type Post struct {
	PostID          int    `json:"post_id"`
	UserID          string `json:"user_id"`
	Content         string `json:"content"`
	Image           string `json:"image_url"`
	CreateTimestamp string `json:"create_timestamp"`
	UpdateTimestamp string `json:"update_timestamp"`
	PostVisibility  string `json:"post_visibility"`
	PostStatus      string `json:"post_status"`
	LikeCount       int    `json:"like_count"`
	CommentCount    int    `json:"comment_count"`
	ReportCount     int    `json:"report_count"`
}

func GetAllPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		row, err := db.Query(`SELECT post_id, user_id_fk, content, image_url, create_timestamp, update_timestamp, post_visibility, post_status, like_count, comment_count, report_count FROM posts WHERE post_visibility = 'public'`)
		if err != nil {
			// debug
			fmt.Println(err)

			http.Error(w, "Error fetch data", 500)
			return
		}

		defer row.Close()

		var post []Post
		for row.Next() {
			var p Post
			err := row.Scan(&p.PostID, &p.UserID, &p.Content, &p.Image, &p.CreateTimestamp, &p.UpdateTimestamp, &p.PostVisibility, &p.PostStatus, &p.LikeCount, &p.CommentCount, &p.ReportCount)
			if err != nil {
				http.Error(w, "something went wrong while read data from Database", 500)
				return
			}
			post = append(post, p)
		}

		if err := row.Err(); err != nil {
			http.Error(w, "something went wrong with row in database", 500)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(post)
	}
}

func GetMyPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(auth.UserContextKey).(*auth.Claims)

		row, err := db.Query(`SELECT post_id, user_id_fk, content, image_url, like_count, comment_count, report_count, create_timestamp, update_timestamp FROM post WHERE user_id_pk = $1`, claims.UserID)

		if err != nil {
			http.Error(w, "Error fetch data", 500)
			return
		}

		defer row.Close()

		var posts []Post

		for row.Next() {
			var p Post
			row.Scan(&p.PostID, &p.UserID, &p.Content, &p.Image, &p.LikeCount, &p.CommentCount, &p.ReportCount, &p.CreateTimestamp, &p.UpdateTimestamp)
			posts = append(posts, p)
		}

		json.NewEncoder(w).Encode(posts)
	}
}
