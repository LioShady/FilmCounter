package web

import (
	"FilmCounter/internal/pkapi"
	"FilmCounter/internal/storage"
	"html/template"
	"log"
	"net/http"
)

type Server struct {
	store     *storage.Store
	logger    *log.Logger
	pkClient  *pkapi.Client
	templates *template.Template
}

func NewServer(store *storage.Store, logger *log.Logger, pkClient *pkapi.Client) *Server {
	tmpl := template.Must(template.ParseGlob("internal/web/templates/*.html"))

	return &Server{store: store, logger: logger, pkClient: pkClient, templates: tmpl}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", s.handleDirectors)
	mux.HandleFunc("GET /directors", s.handleDirectors)
	mux.HandleFunc("GET /import", s.handleImportForm)
	mux.HandleFunc("POST /import/search", s.handleImportSearch)
	mux.HandleFunc("POST /import/add", s.handleImportAdd)
	mux.HandleFunc("POST /films/{id}/delete", s.handleRemove)

	return mux
}
