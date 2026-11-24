package main

import (
	"fmt"
	"sync"
)

// Store keeps all data in memory. ONE mutex protects everything (R8).
type Store struct {
	mu sync.Mutex

	nextOrg, nextUser, nextApp, nextAlert int

	companies    map[string]*Company
	usersByToken map[string]*User // lookup by token = fast auth
	apps         map[string]*App  // by appId
	appsByKey    map[string]*App  // by appKey = fast SDK auth
	alerts       map[string]*Alert
}

func NewStore() *Store {
	return &Store{
		companies:    map[string]*Company{},
		usersByToken: map[string]*User{},
		apps:         map[string]*App{},
		appsByKey:    map[string]*App{},
		alerts:       map[string]*Alert{},
	}
}

// CreateCompanyWithAdmin makes a company and its first admin user.
// It returns a COPY of the user so the caller can use it without the lock.
func (s *Store) CreateCompanyWithAdmin(orgName, name, email string) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextOrg++
	orgID := fmt.Sprintf("org_%d", s.nextOrg)
	s.companies[orgID] = &Company{ID: orgID, Name: orgName}

	return s.addUserLocked(orgID, name, email, "admin")
}

// addUserLocked: the name ends in "Locked" = caller MUST already hold s.mu.
func (s *Store) addUserLocked(orgID, name, email, role string) User {
	s.nextUser++
	u := &User{
		UserID: fmt.Sprintf("usr_%d", s.nextUser),
		OrgID:  orgID,
		Name:   name,
		Email:  email,
		Role:   role,
		Token:  randomHex(),
	}
	s.usersByToken[u.Token] = u
	return *u
}

// UserByToken finds the user who owns this token. Returns a copy.
func (s *Store) UserByToken(token string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.usersByToken[token]
	if !ok {
		return User{}, false
	}
	return *u, true
}

// AppByKey finds the app that owns this app key. Returns a copy.
func (s *Store) AppByKey(key string) (App, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.appsByKey[key]
	if !ok {
		return App{}, false
	}
	return *a, true
}

// AddUser adds a user to an existing company.
func (s *Store) AddUser(orgID, name, email, role string) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addUserLocked(orgID, name, email, role)
}

// AddApp adds an app to a company and gives it a random appKey.
func (s *Store) AddApp(orgID, name, platform string) App {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextApp++
	a := &App{
		AppID:    fmt.Sprintf("app_%d", s.nextApp),
		Name:     name,
		Platform: platform,
		AppKey:   randomHex(),
		OrgID:    orgID,
	}
	s.apps[a.AppID] = a
	s.appsByKey[a.AppKey] = a
	return *a
}

// ReportAlert handles one alert from the SDK.
// It returns the alert (a copy) and created=true if a new alert was made.
// The whole "look for an open alert, then update or create" runs under ONE lock.
func (s *Store) ReportAlert(appID, deviceID, threatType string, severity int) (Alert, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowUTC()

	// 1. Is there an OPEN alert for the same app + device + threat?
	for _, a := range s.alerts {
		if a.AppID == appID && a.DeviceID == deviceID &&
			a.ThreatType == threatType && a.Status == "open" {
			a.Count++        // only count and lastSeen change
			a.LastSeen = now // severity stays as the first one
			return *a, false
		}
	}

	// 2. No open alert: create a new one.
	s.nextAlert++
	a := &Alert{
		ID:         fmt.Sprintf("alr_%d", s.nextAlert),
		AppID:      appID,
		DeviceID:   deviceID,
		ThreatType: threatType,
		Severity:   severity,
		Count:      1,
		Status:     "open",
		FirstSeen:  now,
		LastSeen:   now,
		ResolvedBy: "",
	}
	s.alerts[a.ID] = a
	return *a, true
}