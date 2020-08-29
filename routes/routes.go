package routes

import "net/http"

func serveFrontEnd() {
	fs := http.FileServer(http.Dir("./client/build"))
	http.Handle("/", fs)
}

func RegisterRoutes() {
	serveUserRoutes()
	serveFrontEnd()
}

