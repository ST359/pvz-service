package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ST359/pvz-service/internal/config"
	"github.com/ST359/pvz-service/internal/handler"
	"github.com/ST359/pvz-service/internal/repository"
	"github.com/ST359/pvz-service/internal/service"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
func main() {
	cfg := config.MustLoad()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("error during db initializing: %w", err)
	}
	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services, logger)
	srv := new(Server)
	go func() {
		if err := srv.Run(strconv.Itoa(cfg.Port), handlers.InitRoutes()); err != nil {
			log.Fatalf("error while running server: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Print("Shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Print("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		log.Print("error occured on db connection close: %s", err.Error())
	}
}
