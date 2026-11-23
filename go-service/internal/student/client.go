package student

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Student is the flat student detail payload returned by the Node API.
type Student struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Email              string  `json:"email"`
	SystemAccess       bool    `json:"systemAccess"`
	Phone              *string `json:"phone"`
	Gender             *string `json:"gender"`
	DOB                *string `json:"dob"`
	Class              *string `json:"class"`
	Section            *string `json:"section"`
	Roll               *int    `json:"roll"`
	FatherName         *string `json:"fatherName"`
	FatherPhone        *string `json:"fatherPhone"`
	MotherName         *string `json:"motherName"`
	MotherPhone        *string `json:"motherPhone"`
	GuardianName       *string `json:"guardianName"`
	GuardianPhone      *string `json:"guardianPhone"`
	RelationOfGuardian *string `json:"relationOfGuardian"`
	CurrentAddress     *string `json:"currentAddress"`
	PermanentAddress   *string `json:"permanentAddress"`
	AdmissionDate      *string `json:"admissionDate"`
	ReporterName       *string `json:"reporterName"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("node api error %d: %s", e.StatusCode, e.Message)
}

type Client struct {
	BaseURL  string
	Username string
	Password string
	HTTP     *http.Client
}

func NewClient(baseURL, username, password string) *Client {
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Username: username,
		Password: password,
		HTTP:     &http.Client{Timeout: 10 * time.Second},
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authSession struct {
	cookieHeader string
	csrfToken    string
}

func (c *Client) GetByID(id string) (*Student, error) {
	session, err := c.login()
	if err != nil {
		return nil, err
	}
	return c.fetchStudent(id, session)
}

func (c *Client) login() (authSession, error) {
	body, err := json.Marshal(loginRequest{
		Username: c.Username,
		Password: c.Password,
	})
	if err != nil {
		return authSession{}, fmt.Errorf("encode login body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return authSession{}, fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return authSession{}, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return authSession{}, fmt.Errorf("read login response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return authSession{}, apiErrorFromBody(resp.StatusCode, respBody, "login failed")
	}

	cookies := cookieMap(resp.Cookies())
	accessToken, okAccess := cookies["accessToken"]
	refreshToken, okRefresh := cookies["refreshToken"]
	csrfToken, okCSRF := cookies["csrfToken"]
	if !okAccess || !okRefresh || !okCSRF || accessToken == "" || refreshToken == "" || csrfToken == "" {
		return authSession{}, fmt.Errorf("login succeeded but auth cookies were missing")
	}

	return authSession{
		cookieHeader: fmt.Sprintf("accessToken=%s; refreshToken=%s; csrfToken=%s", accessToken, refreshToken, csrfToken),
		csrfToken:    csrfToken,
	}, nil
}

func (c *Client) fetchStudent(id string, session authSession) (*Student, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/v1/students/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("create student request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", session.cookieHeader)
	req.Header.Set("X-CSRF-Token", session.csrfToken)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("student request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read student response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fallback := "failed to fetch student"
		if resp.StatusCode == http.StatusNotFound {
			fallback = "Student not found"
		}
		return nil, apiErrorFromBody(resp.StatusCode, respBody, fallback)
	}

	var student Student
	if err := json.Unmarshal(respBody, &student); err != nil {
		return nil, fmt.Errorf("decode student response: %w", err)
	}
	return &student, nil
}

func cookieMap(cookies []*http.Cookie) map[string]string {
	out := make(map[string]string, len(cookies))
	for _, c := range cookies {
		out[c.Name] = c.Value
	}
	return out
}

func apiErrorFromBody(status int, body []byte, fallback string) *APIError {
	return &APIError{
		StatusCode: status,
		Message:    extractErrorMessage(body, fallback),
	}
}

func extractErrorMessage(body []byte, fallback string) string {
	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if payload.Error != "" {
			return payload.Error
		}
		if payload.Message != "" {
			return payload.Message
		}
	}
	if trimmed := strings.TrimSpace(string(body)); trimmed != "" {
		return trimmed
	}
	return fallback
}
