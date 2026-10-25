package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	d "acm.stupay.com/internal/database"
	ws "acm.stupay.com/internal/wallet"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-fuego/fuego"
	"github.com/go-playground/validator/v10"
	gocache "github.com/patrickmn/go-cache"
	"github.com/rs/cors"
)

type Server struct {
	server         *fuego.Server
	logger         *slog.Logger
	DB             *d.DataService
	ws             *ws.WalletService
	sessionManager *scs.SessionManager
	cache          *gocache.Cache
}

type NewServerArgs struct {
	Logger *slog.Logger
	DB     *sql.DB
}

func (s *Server) Run() {
	s.logger.Info("Starting server...")
	err := s.server.Run()
	if err != nil {
		s.logger.Error(err.Error())
	}
}

func InitServer(args NewServerArgs) Server {
	if args.Logger == nil {
		args.Logger = slog.Default()
	}
	if args.DB == nil {
		args.Logger.Error("No db connection given")
		panic("must provide db connection")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		args.Logger.Warn("FRONTEND_URL not set; using localhost only")
		frontendURL = "http://localhost:5173"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port if not specified
	}
	host := os.Getenv("HOST") // Optional
	addr := ":" + port
	if host != "" {
		addr = host + ":" + port
	}

	DB := d.NewDataService(d.NewDataServiceArgs{
		Db:        args.DB,
		Logger:    args.Logger,
		Validator: validator.New(),
	})

	// Cache items expire after 15 minutes,
	// and the cache is cleaned up every 30 minutes.
	appCache := gocache.New(15*time.Minute, 30*time.Minute)

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(args.DB)
	sessionManager.Lifetime = 999 * time.Hour
	sessionManager.IdleTimeout = 120 * time.Hour
	sessionManager.Cookie.Name = "session_id"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = true
	sessionManager.Cookie.Partitioned = true

	server := fuego.NewServer(
		fuego.WithAddr(addr),
		fuego.WithGlobalMiddlewares(cors.New(cors.Options{
			AllowedOrigins:   []string{frontendURL, "http://localhost:5173"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
		}).Handler),
	)

	ws := ws.NewWalletService(ws.NewWalletServiceArgs{
		Logger:   args.Logger,
		Database: DB,
	})

	return Server{
		logger:         args.Logger,
		server:         server,
		DB:             DB,
		sessionManager: sessionManager,
		cache:          appCache,
		ws:             ws,
	}
}
