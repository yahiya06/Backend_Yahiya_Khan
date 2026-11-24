package main

import (
	"net/http"
	"strings"
)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgName string `json:"orgName"`
		Name    string `json:"name"`
		Email   string `json:"email"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	req.OrgName = strings.TrimSpace(req.OrgName)
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	if req.OrgName == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "orgName and name must not be empty")
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "email must contain @")
		return
	}

	user := s.store.CreateCompanyWithAdmin(req.OrgName, req.Name, req.Email)
	writeJSON(w, http.StatusCreated, user)
}
