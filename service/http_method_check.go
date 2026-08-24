package service

import (
	"net/http"
)

func HttpMethodPost(r *http.Request) bool {
	return r.Method != http.MethodPost
}
