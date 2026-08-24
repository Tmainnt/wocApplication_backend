package pdb

import (
	"database/sql"
	"net/http"
)

func WriteBack(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
