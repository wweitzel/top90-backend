package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wweitzel/top90/internal/api/handlers"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
)

type Server struct {
	dao      db.Top90DAO
	s3Client s3.S3Client
	router   *mux.Router
	config   config.Config
}

func NewServer(dao db.Top90DAO, s3Client s3.S3Client, config config.Config) *Server {
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
	welcomeHandler := handlers.WelcomeHandler{}
	fixturesHandler := handlers.NewFixtureHandler(s.dao)
	goalHandler := handlers.NewGoalHandler(s.dao, s.s3Client, s.config.AwsBucketName)
	leagueHandler := handlers.NewLeagueHandler(s.dao)
	messageHandler := handlers.NewMessageHandler(s.dao, s.s3Client, s.config.AwsBucketName)
	teamsHandler := handlers.NewTeamHandler(s.dao)

	s.router.HandleFunc("/", welcomeHandler.GetWelcome)
	s.router.HandleFunc("/fixtures", fixturesHandler.GetFixtures)
	s.router.HandleFunc("/fixtures/{id}", fixturesHandler.GetFixture)
	s.router.HandleFunc("/goals", goalHandler.GetGoals)
	s.router.HandleFunc("/goals/{id}", goalHandler.GetGoal)
	s.router.HandleFunc("/leagues", leagueHandler.GetLeagues)
	s.router.HandleFunc("/message_preview/{id}", messageHandler.GetPreview)
	s.router.HandleFunc("/teams", teamsHandler.GetTeams)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	log.Printf("%s %s", r.Method, r.URL)
	s.router.ServeHTTP(w, r)
}
