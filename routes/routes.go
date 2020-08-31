package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"path/filepath"
)

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

func serveCatchAllAPIRoutes(r *mux.Router) {
	r.PathPrefix("/api").HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		http.Error(w, "Route not found", http.StatusNotFound)
	})
}

func serveFrontEnd(r *mux.Router) {
	r.PathPrefix("/").HandlerFunc(serveSPA)
}

func RegisterRoutes(r *mux.Router) {
	serveUserRoutes(r)
	serveCatchAllAPIRoutes(r)
	serveFrontEnd(r)
}

