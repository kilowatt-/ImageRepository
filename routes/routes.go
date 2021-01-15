package routes

import (
	"github.com/gorilla/mux"
	"github.com/kilowatt-/ImageRepository/routes/images"
	"github.com/kilowatt-/ImageRepository/routes/users"
	"net/http"
)

func serveCatchAllAPIRoutes(r *mux.Router) {
	r.PathPrefix("/api").HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		http.Error(w, "Route not found", http.StatusNotFound)
	})
}


func RegisterRoutes(r *mux.Router) {
	users.ServeUserRoutes(r.PathPrefix("/api/users").Subrouter())
	images.ServeImageRoutes(r.PathPrefix("/api/images").Subrouter())
	serveCatchAllAPIRoutes(r)
}

