package main

import (
	"encoding/json"
	"net/http"
	"strconv"
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

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	caller, ok := s.authAdmin(w, r) // 401, then 403
	if !ok {
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if !decodeJSON(w, r, &req) { // 400
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name must not be empty")
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "email must contain @")
		return
	}
	if req.Role != "admin" && req.Role != "viewer" {
		writeError(w, http.StatusBadRequest, `role must be "admin" or "viewer"`)
		return
	}

	// The company comes from the CALLER, never from the request body (R6).
	user := s.store.AddUser(caller.OrgID, req.Name, req.Email, req.Role)
	writeJSON(w, http.StatusCreated, user)
}

func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	caller, ok := s.authAdmin(w, r) // 401, then 403
	if !ok {
		return
	}

	var req struct {
		Name     string `json:"name"`
		Platform string `json:"platform"`
	}
	if !decodeJSON(w, r, &req) { // 400
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name must not be empty")
		return
	}
	if req.Platform != "android" && req.Platform != "ios" {
		writeError(w, http.StatusBadRequest, `platform must be "android" or "ios"`)
		return
	}

	app := s.store.AddApp(caller.OrgID, req.Name, req.Platform)
	writeJSON(w, http.StatusCreated, app)
}

func (s *Server) handleReportAlert(w http.ResponseWriter, r *http.Request) {
	app, ok := s.authApp(w, r) // 401 for missing or wrong app key
	if !ok {
		return
	}

	// severity is a pointer so we can tell "missing" from "0".
	// json.Number keeps the raw number so we can reject 4.5 or "5".
	var req struct {
		DeviceID   string       `json:"deviceId"`
		ThreatType string       `json:"threatType"`
		Severity   *json.Number `json:"severity"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.ThreatType = strings.TrimSpace(req.ThreatType)

	if req.DeviceID == "" || req.ThreatType == "" {
		writeError(w, http.StatusBadRequest, "deviceId and threatType must not be empty")
		return
	}
	if req.Severity == nil {
		writeError(w, http.StatusBadRequest, "severity is required")
		return
	}
	// Atoi fails for 4.5, 1e1, and anything that is not a whole number.
	severity, err := strconv.Atoi(req.Severity.String())
	if err != nil || severity < 1 || severity > 5 {
		writeError(w, http.StatusBadRequest, "severity must be a whole number from 1 to 5")
		return
	}

	alert, created := s.store.ReportAlert(app.AppID, req.DeviceID, req.ThreatType, severity)
	if created {
		writeJSON(w, http.StatusCreated, alert) // 201
	} else {
		writeJSON(w, http.StatusOK, alert) // 200
	}
}
