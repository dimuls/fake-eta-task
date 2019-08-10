package web

import (
	"context"
	"net/http"

	"github.com/dimuls/fake-eta-task/entity"
	"github.com/sirupsen/logrus"
)

type WAPIClient interface {
	Cars(coord *entity.Coordinate, limit int) (
		[]*entity.Coordinate, error)
	Predicts(target *entity.Coordinate,
		sources []*entity.Coordinate) ([]int, error)
}

type Server struct {
	addr string
	wapi WAPIClient
	serv *http.Server
	log  *logrus.Entry
}

func NewServer(addr string, wapi WAPIClient) *Server {
	return &Server{
		addr: addr,
		wapi: wapi,
		log:  logrus.WithField("subsystem", "http_server"),
	}
}

func (s *Server) Start() error {
	s.serv = &http.Server{
		Addr:    s.addr,
		Handler: &Handler{wapi: s.wapi, log: s.log},
	}

	err := s.serv.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	return s.serv.Shutdown(context.TODO())
}
