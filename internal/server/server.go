package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/s3"
)

type Server struct {
	dao      db.Top90DAO
	s3Client s3.S3Client
	router   *mux.Router
	config   top90.Config
}

func NewServer(dao db.Top90DAO, s3Client s3.S3Client, config top90.Config) *Server {
	s := &Server{
		dao:      dao,
		s3Client: s3Client,
		router:   mux.NewRouter(),
		config:   config,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/", s.GetApiInfoHandler)
	s.router.HandleFunc("/fixtures", s.GetFixturesHandler)
	s.router.HandleFunc("/fixtures/{id}", s.GetFixtureHandler)
	s.router.HandleFunc("/goals", s.GetGoalsHandler)
	s.router.HandleFunc("/goals/{id}", s.GetGoalHandler)
	s.router.HandleFunc("/leagues", s.GetLeaguesHandler)
	s.router.HandleFunc("/teams", s.GetTeamsHandler)
	s.router.HandleFunc("/message_preview/{id}", s.MessagePreviewHandler)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	log.Printf("%s %s", r.Method, r.URL)
	s.router.ServeHTTP(w, r)
}
