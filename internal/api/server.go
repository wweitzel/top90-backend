package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wweitzel/top90/internal/api/handlers"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db/dao"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func metricsMiddleware(next http.HandlerFunc, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestsTotal.WithLabelValues(endpoint, r.Method).Inc()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		requestDuration.WithLabelValues(endpoint, r.Method).Observe(duration.Seconds())
	}
}

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"endpoint", "method"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)
)

type Server struct {
	dao      dao.Top90DAO
	s3Client s3.S3Client
	router   *chi.Mux
	config   config.Config
	logger   *slog.Logger
	metrics  *newrelic.Application
}

func NewServer(dao dao.Top90DAO, s3Client s3.S3Client, config config.Config, logger *slog.Logger, metrics *newrelic.Application) *Server {
	s := &Server{
		dao:      dao,
		s3Client: s3Client,
		router:   chi.NewRouter(),
		config:   config,
		logger:   logger,
		metrics:  metrics,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	welcomeHandler := handlers.WelcomeHandler{}
	authHandler := handlers.AuthHandler{}
	optionsHandler := handlers.OptionsHandler{}
	fixturesHandler := handlers.NewFixtureHandler(s.dao)
	goalHandler := handlers.NewGoalHandler(s.dao, s.s3Client, s.config.AwsBucketName)
	leagueHandler := handlers.NewLeagueHandler(s.dao)
	messageHandler := handlers.NewMessageHandler(s.dao, s.s3Client, s.config.AwsBucketName)
	playerHandler := handlers.NewPlayerHandler(s.dao)
	teamsHandler := handlers.NewTeamHandler(s.dao)
	logoHandler := handlers.NewLogoHandler()

	// Helper to wrap and register routes
	register := func(method, pattern string, handler http.HandlerFunc) {
		_, h := newrelic.WrapHandleFunc(s.metrics, pattern, handler)
		switch method {
		case "GET":
			s.router.Get(pattern, h)
		case "POST":
			s.router.Post(pattern, h)
		case "DELETE":
			s.router.Delete(pattern, h)
		case "OPTIONS":
			s.router.Options(pattern, h)
		}
	}

	register("GET", "/", welcomeHandler.GetWelcome)
	register("GET", "/fixtures", fixturesHandler.GetFixtures)
	register("GET", "/fixtures/{id}", fixturesHandler.GetFixture)
	register("GET", "/goals", metricsMiddleware(goalHandler.GetGoals, "goals"))
	register("GET", "/goals/{id}", goalHandler.GetGoal)
	register("DELETE", "/goals/{id}", goalHandler.DeleteGoal)
	register("OPTIONS", "/goals/{id}", optionsHandler.Default)
	register("GET", "/leagues", leagueHandler.GetLeagues)
	register("GET", "/message_preview/{id}", messageHandler.GetPreview)
	register("GET", "/players", playerHandler.SearchPlayers)
	register("GET", "/teams", teamsHandler.GetTeams)
	register("GET", "/logo/{id}", logoHandler.GetLogo)
	register("POST", "/login", authHandler.Login)
	register("OPTIONS", "/login", optionsHandler.Default)
	register("POST", "/logout", authHandler.Logout)
	register("OPTIONS", "/logout", optionsHandler.Default)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", os.Getenv("TOP90_ORIGIN"))
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	s.logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))
	s.router.ServeHTTP(w, r)
}
