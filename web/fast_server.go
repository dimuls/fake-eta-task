package web

import (
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type FastServer struct {
	addr string
	wapi WAPIClient
	serv *fasthttp.Server
	log  *logrus.Entry
}

func NewFastServer(addr string, wapi WAPIClient) *Server {
	return &Server{
		addr: addr,
		wapi: wapi,
		log:  logrus.WithField("subsystem", "http_server"),
	}
}

func (s *FastServer) Start() error {
	s.serv = &fasthttp.Server{
		Handler: (&FastHandler{wapi: s.wapi, log: s.log}).Handle,
	}
	return s.serv.ListenAndServe(s.addr)
}

func (s *FastServer) Stop() error {
	return s.serv.Shutdown()
}
