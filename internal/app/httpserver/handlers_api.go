package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handlePrograms(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.mgr.StatusAll())
}

func (s *Server) handleProgram(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	status, err := s.mgr.Status(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleProgramStart(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.mgr.Start(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (s *Server) handleProgramStop(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.mgr.Stop(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) handleProgramRestart(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.mgr.Restart(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "restarted"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
