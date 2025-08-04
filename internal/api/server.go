package api

import (
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
)

type Server struct {
	router *mux.Router
	logger *slog.Logger
}

func NewServer(logger *slog.Logger) *Server {
	slog.SetDefault(logger)

	router := mux.NewRouter()
	logger.Info("Server created. Returning Server.")

	return &Server{router: router, logger: logger}
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.router.ServeHTTP(writer, request)
}
