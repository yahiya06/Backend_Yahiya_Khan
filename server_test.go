package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// ---------- helpers ----------

// do sends one request straight to the handler (no real server needed).
func do(t *testing.T, h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decode reads the JSON response body into a map.
func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("response is not JSON: %v (%s)", err, rec.Body.String())
	}
	return m
}

// setup registers Acme, adds a viewer and an app, and sends one alert.
// It returns the tokens, the app key and the alert id.
type fixture struct {
	h                       http.Handler
	adminTok, viewerTok     string
	appKey, alertID         string
}

func setup(t *testing.T) fixture {
	t.Helper()
	h := NewServer().Routes() // fresh empty server for every test

	rec := do(t, h, "POST", "/register", `{"orgName":"Acme","name":"Asha","email":"asha@acme.test"}`, nil)
	adminTok := decode(t, rec)["token"].(string)

	rec = do(t, h, "POST", "/users", `{"name":"Vik","email":"vik@acme.test","role":"viewer"}`,
		map[string]string{"Authorization": "Bearer " + adminTok})
	viewerTok := decode(t, rec)["token"].(string)

	rec = do(t, h, "POST", "/apps", `{"name":"Acme Pay","platform":"android"}`,
		map[string]string{"Authorization": "Bearer " + adminTok})
	appKey := decode(t, rec)["appKey"].(string)

	rec = do(t, h, "POST", "/alerts", `{"deviceId":"d1","threatType":"ROOT_DETECTED","severity":5}`,
		map[string]string{"X-App-Key": appKey})
	alertID := decode(t, rec)["id"].(string)

	return fixture{h, adminTok, viewerTok, appKey, alertID}
}

// ---------- the 3 required tests ----------

// 1. A viewer gets 403 when resolving an alert.
func TestViewerCannotResolve(t *testing.T) {
	f := setup(t)

	rec := do(t, f.h, "POST", "/alerts/"+f.alertID+"/resolve", "",
		map[string]string{"Authorization": "Bearer " + f.viewerTok})

	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d (%s)", rec.Code, rec.Body.String())
	}
}

// 2. A user from another company gets 404 for an alert.
func TestOtherCompanyGets404(t *testing.T) {
	f := setup(t)

	rec := do(t, f.h, "POST", "/register", `{"orgName":"Beta","name":"Ravi","email":"ravi@beta.test"}`, nil)
	betaTok := decode(t, rec)["token"].(string)

	rec = do(t, f.h, "GET", "/alerts/"+f.alertID, "",
		map[string]string{"Authorization": "Bearer " + betaTok})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d (%s)", rec.Code, rec.Body.String())
	}
}

// 3. 50 identical alerts at the same time give ONE alert with count 50.
func TestConcurrentAlertsGiveOneAlert(t *testing.T) {
	f := setup(t)
	body := `{"deviceId":"d2","threatType":"FRIDA_DETECTED","severity":4}`

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			do(t, f.h, "POST", "/alerts", body, map[string]string{"X-App-Key": f.appKey})
		}()
	}
	wg.Wait() // wait until all 50 requests are finished

	rec := do(t, f.h, "GET", "/alerts?status=open", "",
		map[string]string{"Authorization": "Bearer " + f.adminTok})
	alerts := decode(t, rec)["alerts"].([]any)

	// Count alerts for d2 (setup also made one alert for d1).
	found := 0
	for _, item := range alerts {
		a := item.(map[string]any)
		if a["deviceId"] == "d2" {
			found++
			if a["count"].(float64) != 50 {
				t.Fatalf("want count 50, got %v", a["count"])
			}
		}
	}
	if found != 1 {
		t.Fatalf("want exactly 1 alert for d2, got %d", found)
	}
}