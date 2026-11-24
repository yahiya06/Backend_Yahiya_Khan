package main

import (
	"net/http"
	"strings"
)

// authUser checks "Authorization: Bearer <token>".
// On failure it sends 401 and returns false, so the handler just returns.
func (s *Server) authUser(w http.ResponseWriter, r *http.Request) (User, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		writeError(w, http.StatusUnauthorized, "missing or invalid token")
		return User{}, false
	}
	user, found := s.store.UserByToken(token)
	if !found {
		writeError(w, http.StatusUnauthorized, "missing or invalid token")
		return User{}, false
	}
	return user, true
}

// authAdmin = authUser + role must be admin (401 first, then 403).
func (s *Server) authAdmin(w http.ResponseWriter, r *http.Request) (User, bool) {
	user, ok := s.authUser(w, r)
	if !ok {
		return User{}, false
	}
	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "only an admin can do this")
		return User{}, false
	}
	return user, true
}

// authApp checks "X-App-Key: <appKey>" for the SDK.
func (s *Server) authApp(w http.ResponseWriter, r *http.Request) (App, bool) {
	key := r.Header.Get("X-App-Key")
	if key == "" {
		writeError(w, http.StatusUnauthorized, "missing or invalid app key")
		return App{}, false
	}
	app, found := s.store.AppByKey(key)
	if !found {
		writeError(w, http.StatusUnauthorized, "missing or invalid app key")
		return App{}, false
	}
	return app, true
}
