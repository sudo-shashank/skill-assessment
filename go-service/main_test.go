package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"skill-assessment/go-service/internal/student"
)

func writeLoginCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "accessToken", Value: "access-token"})
	http.SetCookie(w, &http.Cookie{Name: "refreshToken", Value: "refresh-token"})
	http.SetCookie(w, &http.Cookie{Name: "csrfToken", Value: "csrf-token"})
}

func newReportMux(client *student.Client) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/students/{id}/report", reportHandler(client))
	return mux
}

func TestReportHandlerSuccess(t *testing.T) {
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/login":
			writeLoginCookies(w)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Admin"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/students/42":
			if got := r.Header.Get("X-CSRF-Token"); got != "csrf-token" {
				t.Fatalf("expected csrf header csrf-token, got %q", got)
			}
			if !strings.Contains(r.Header.Get("Cookie"), "accessToken=access-token") {
				t.Fatalf("expected accessToken cookie, got %q", r.Header.Get("Cookie"))
			}
			phone := "555-0100"
			className := "Grade 10"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(student.Student{
				ID:           42,
				Name:         "Jane Doe",
				Email:        "jane@example.com",
				SystemAccess: true,
				Phone:        &phone,
				Class:        &className,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer node.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/students/42/report", nil)
	rr := httptest.NewRecorder()
	newReportMux(student.NewClient(node.URL, "admin@school-admin.com", "password")).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Fatalf("expected application/pdf, got %q", ct)
	}
	if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(cd, "student-42-report.pdf") {
		t.Fatalf("unexpected Content-Disposition: %q", cd)
	}
	if !strings.HasPrefix(rr.Body.String(), "%PDF") {
		t.Fatalf("expected PDF body to start with %%PDF, got %q", rr.Body.String()[:min(20, rr.Body.Len())])
	}
}

func TestReportHandlerUpstreamNotFound(t *testing.T) {
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/login":
			writeLoginCookies(w)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":1}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/students/"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Student not found"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer node.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/students/999/report", nil)
	rr := httptest.NewRecorder()
	newReportMux(student.NewClient(node.URL, "admin@school-admin.com", "password")).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Student not found") {
		t.Fatalf("expected not found message, got %s", rr.Body.String())
	}
}

func TestReportHandlerInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/students/abc/report", nil)
	rr := httptest.NewRecorder()
	newReportMux(student.NewClient("http://example.invalid", "admin", "password")).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "student id must be a number") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestReportHandlerLoginFailure(t *testing.T) {
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/login" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer node.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/students/42/report", nil)
	rr := httptest.NewRecorder()
	newReportMux(student.NewClient(node.URL, "admin@school-admin.com", "bad-password")).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "failed to authenticate with node api") {
		t.Fatalf("expected auth failure message, got %s", rr.Body.String())
	}
}
