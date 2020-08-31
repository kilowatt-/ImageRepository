package routes

import (
	"net/http"
	"os"
	"path/filepath"
)

func allowCookiesInHeader(w *http.ResponseWriter) {
	allowedOrigins, _ := os.LookupEnv("ALLOWED_CORS_ORIGINS")

	(*w).Header().Set("Access-Control-Allow-Origin",allowedOrigins)
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Set-Cookie, *")
}

func serveSPA(w http.ResponseWriter, r *http.Request) {
	staticPath := "./client/build"
	indexPath := "index.html"

	path, err := filepath.Abs(r.URL.Path)

	if err != nil {
		http.Error(w, "Bad path sent", http.StatusBadRequest)
	} else {
		path = filepath.Join(staticPath, path)

		_, err = os.Stat(path)

		if os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(staticPath, indexPath))
		} else if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		} else {
			http.FileServer(http.Dir(staticPath)).ServeHTTP(w, r)
		}
	}
}

func serveFrontEnd() {
	http.HandleFunc("/", serveSPA)
}

func RegisterRoutes() {
	serveUserRoutes()
	serveFrontEnd()
}

