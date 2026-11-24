package main

// Company is one customer. Created by /register.
type Company struct {
	ID   string
	Name string
}

// User JSON field names must match the paper exactly (R3).
type User struct {
	UserID string `json:"userId"`
	OrgID  string `json:"orgId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"` // "admin" or "viewer"
	Token  string `json:"token"`
}

type App struct {
	AppID    string `json:"appId"`
	Name     string `json:"name"`
	Platform string `json:"platform"` // "android" or "ios"
	AppKey   string `json:"appKey"`
	OrgID    string `json:"-"` // internal only, never sent to the client
}

type Alert struct {
	ID         string `json:"id"`
	AppID      string `json:"appId"`
	DeviceID   string `json:"deviceId"`
	ThreatType string `json:"threatType"`
	Severity   int    `json:"severity"`
	Count      int    `json:"count"`
	Status     string `json:"status"` // "open" or "resolved"
	FirstSeen  string `json:"firstSeen"`
	LastSeen   string `json:"lastSeen"`
	ResolvedBy string `json:"resolvedBy"`
}
