package server

import (
	"context"
	"github.com/educ-educ/gateway/internal/pkg/common"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	http.Server
	logger common.Logger
	router *gin.Engine
}

func NewServer(logger common.Logger, router *gin.Engine, addr string) *Server {
	serv := &Server{
		logger: logger,
		router: router,
	}

	serv.configure(addr)

	return serv
}

func (s *Server) Start() error {
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("Server started")

		err := s.ListenAndServe()
		if err != nil {
			errChan <- err
			return
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	signal.Notify(sigChan, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		s.logger.Info("Received terminate, graceful shutdown. Signal:", sig)
		tc, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelFunc()

		_ = s.Shutdown(tc)
	case err := <-errChan:
		return err
	}

	return nil
}

func (s *Server) configure(addr string) {
	s.Addr = addr
	s.Handler = s.router
	s.IdleTimeout = 5 * time.Second
	s.ReadTimeout = 5 * time.Second
	s.WriteTimeout = 5 * time.Second
}
