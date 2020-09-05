package routes

import (
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/routes/images"
	"github.com/kilowatt-/ImageRepository/routes/users"
	"net/http"
	"os"
	"path/filepath"
)
func SendInternalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
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
			SendInternalServerError(w)
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
	users.ServeUserRoutes(r.PathPrefix("/api/users").Subrouter())
	images.ServeImageRoutes(r.PathPrefix("/api/images").Subrouter())
	serveCatchAllAPIRoutes(r)
	serveFrontEnd(r)
}

