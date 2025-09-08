package api

import (
	"database/sql"
	apiV1 "jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/services"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
	logger *slog.Logger
}

func NewServer(database *sql.DB, logger *slog.Logger) *Server {
	slog.SetDefault(logger)

	companyRepository := repositories.NewCompanyRepository(database)
	companyService := services.NewCompanyService(companyRepository)
	companyHandler := apiV1.NewCompanyHandler(companyService)

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/company/new", companyHandler.CreateCompany).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/company/get/id/{id}", companyHandler.GetCompanyById).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/company/get/name/{name}", companyHandler.GetCompaniesByName).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/company/get/all", companyHandler.GetAllCompanies).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/company/update", companyHandler.UpdateCompany).Methods(http.MethodPost)

	logger.Info("Server created. Returning Server.")
	return &Server{router: router, logger: logger}
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.router.ServeHTTP(writer, request)
}
