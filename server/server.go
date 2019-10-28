package server

import (
	logger "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Server struct {
	http.Server
	Logger *logger.Entry
}

func (s *Server) Start() {
	s.Logger.Info("server starting")
	if err := s.ListenAndServe(); err != nil {
		s.Logger.Fatalln(err)
	}
}

func NewServer(addr string) *Server {
	log := logger.New()
	log.SetLevel(logger.DebugLevel)
	log.SetOutput(os.Stdout)
	entry := log.WithFields(logger.Fields{
		"addr":addr,
		"package":"server",
	})
	return &Server{
		Logger: entry,
		Server: http.Server{Addr: addr},
	}
}

func (s *Server) RegisterHandleFunc(pattern string, handler http.Handler) {
	http.Handle(pattern,handler)
}


