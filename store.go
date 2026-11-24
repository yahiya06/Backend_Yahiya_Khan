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
