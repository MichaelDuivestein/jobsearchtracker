package api

import (
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
)

type Server struct {
	router *mux.Router
}

func NewServer() *Server {
	router := mux.NewRouter()

	slog.Info("Server created. Returning Server.")
	return &Server{router: router}
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.router.ServeHTTP(writer, request)
}
