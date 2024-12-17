package server

import (
	"encoding/json"
	"net/http"

	"main.go/cache"
)

type Server struct {
	cache *cache.Cache
}

func NewServer(c *cache.Cache) *Server {
	return &Server{cache: c}
}

func (s *Server) Start(port string) {
	http.HandleFunc("/order", s.getOrderByID)
	http.ListenAndServe(":"+port, nil)
}

func (s *Server) getOrderByID(w http.ResponseWriter, r *http.Request) {
	orderUID := r.URL.Query().Get("id")
	if orderUID == "" {
		http.Error(w, "Требуется ID", http.StatusBadRequest)
		return
	}

	order, exists := s.cache.Get(orderUID)
	if !exists {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
