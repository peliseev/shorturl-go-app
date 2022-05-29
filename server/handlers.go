package server

import (
	"log"
	"net/http"
)

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[1:]
	if p == "" || len(p) != 5 {
		notFound(w, r)
		return
	}
	su, err := s.urlService.GetOriginUrl(r.Context(), p)
	if err != nil {
		log.Print(err)
		notFound(w, r)
		return
	}
	http.Redirect(w, r, su.OriginURL, http.StatusFound)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("Invalid request: %q", r.URL.Path)
	http.NotFound(w, r)
	return
}
