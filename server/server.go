package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/peliseev/shorturl-go-app/domain"
	"github.com/peliseev/shorturl-go-app/mongo"
	driver "go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	server     *http.Server
	db         *driver.Client
	urlService domain.ShortURLService
}

func NewServer(db *driver.Client, sus *mongo.ShortURLService) *Server {
	s := &Server{
		server: &http.Server{
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
			IdleTimeout:  5 * time.Second,
		},
		db:         db,
		urlService: sus,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handler)
	s.server.Handler = mux

	return s
}

func (s *Server) Run(port string) error {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	s.server.Addr = port
	log.Printf("Server starting on %s", port)
	return s.server.ListenAndServe()
}
