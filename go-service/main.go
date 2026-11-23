package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"skill-assessment/go-service/internal/config"
	"skill-assessment/go-service/internal/report"
	"skill-assessment/go-service/internal/student"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	client := student.NewClient(cfg.NodeAPIURL, cfg.NodeAPIUsername, cfg.NodeAPIPassword)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/students/{id}/report", reportHandler(client))

	addr := ":" + cfg.Port
	log.Printf("Go PDF report service listening on %s (node api: %s)", addr, cfg.NodeAPIURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func reportHandler(client *student.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := strconv.Atoi(id); err != nil {
			writeJSONError(w, http.StatusBadRequest, "student id must be a number")
			return
		}

		s, err := client.GetByID(id)
		if err != nil {
			status, message := mapUpstreamError(err)
			writeJSONError(w, status, message)
			return
		}

		pdfBytes, err := report.Generate(s)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to generate pdf")
			return
		}

		filename := fmt.Sprintf("student-%s-report.pdf", id)
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(pdfBytes)
	}
}

func mapUpstreamError(err error) (int, string) {
	var apiErr *student.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusNotFound:
			return http.StatusNotFound, apiErr.Message
		case http.StatusUnauthorized, http.StatusForbidden:
			return http.StatusBadGateway, "failed to authenticate with node api"
		default:
			return http.StatusBadGateway, "upstream request failed"
		}
	}
	return http.StatusBadGateway, "upstream request failed"
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
