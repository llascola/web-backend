package rest

import (
	"context"
	"net/http"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	srv *http.Server
}

func NewServer(handler http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Addr:    ":8080",
			Handler: handler,
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
