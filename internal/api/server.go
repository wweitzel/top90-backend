package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wweitzel/top90/internal/api/handlers"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
)

type Server struct {
	dao      db.Top90DAO
	s3Client s3.S3Client
	router   *chi.Mux
	config   config.Config
	logger   *slog.Logger
}

func NewServer(dao db.Top90DAO, s3Client s3.S3Client, config config.Config, logger *slog.Logger) *Server {
	s := &Server{
		dao:      dao,
		s3Client: s3Client,
		router:   chi.NewRouter(),
		config:   config,
		logger:   logger,
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

	s.router.Get("/", welcomeHandler.GetWelcome)
	s.router.Get("/fixtures", fixturesHandler.GetFixtures)
	s.router.Get("/fixtures/{id}", fixturesHandler.GetFixture)
	s.router.Get("/goals", goalHandler.GetGoals)
	s.router.Get("/goals/{id}", goalHandler.GetGoal)
	s.router.Get("/leagues", leagueHandler.GetLeagues)
	s.router.Get("/message_preview/{id}", messageHandler.GetPreview)
	s.router.Get("/teams", teamsHandler.GetTeams)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	s.logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))
	s.router.ServeHTTP(w, r)
}
