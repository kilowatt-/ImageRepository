package routes

import "net/http"

func serveStatic() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
}

func RegisterRoutes() {
	serveStatic()
}

