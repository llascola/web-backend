package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	srv *http.Server
}

func NewServer(router *gin.Engine) *Server {
	return &Server{
		srv: &http.Server{
			Addr:              ":8001",
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	g, errCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := s.srv.ListenAndServe()
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-errCtx.Done()
		if err := s.Shutdown(context.Background()); err != nil {
			return err
		}
		return nil
	})

	return g.Wait()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
