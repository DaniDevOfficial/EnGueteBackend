package request

import "net/http"

// TODO: do later

func GetFromURL(key string, r *http.Request) string {
	return r.URL.Query().Get(key)
}
