package access

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// ServerPing pings the server
func (repo *Repository) ServerPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type Result struct {
		Status     int    `json:"status"`
		StatusText string `json:"status_text"`
	}
	resp := Result{
		Status:     http.StatusOK,
		StatusText: "Welcome to Data Access Layer",
	}

	_ = json.NewEncoder(w).Encode(resp)
	return

}

// GetServiceSeverity statistics of  a services
func (repo *Repository) GetServiceSeverity(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")

	type Result struct {
		Status          int         `json:"status"`
		StatusText      string      `json:"status_text"`
		Service         interface{} `json:"services,omitempty"`
		ServiceSeverity interface{} `json:"services_severity,omitempty"`
	}

	type ServiceLogs struct {
		ServiceName string `json:"severity_name"  db:"severity_name"`
		Severity    string `json:"service_severity"  db:"service_severity"`
		Count       int    `json:"count"  db:"count"`
	}

	type ServiceSeverity struct {
		ServiceName string `json:"severity_name"  db:"severity_name"`
		Severity    string `json:"service_severity"  db:"service_severity"`
		Count       int    `json:"count"  db:"count"`
	}

	servName := r.FormValue("service-name")
	severity := r.FormValue("severity")

	if len(servName) < 1 {
		resp := Result{
			Status:     http.StatusBadRequest,
			StatusText: "Service name not supplied",
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if len(severity) < 1 {
		resp := Result{
			Status:     http.StatusBadRequest,
			StatusText: "severity not supplied",
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Service severity table check
	s := `SELECT COALESCE (SUM(x.COUNT),0) AS cc FROM lms.service_severity x
		  WHERE (x.service_name = ?) AND (x.severity = ?);`

	servCount := 0
	err = repo.App.Conn.DB.QueryRowContext(ctx, s, servName, severity).Scan(&servCount)
	if err != nil {
		resp := Result{
			Status:          http.StatusInternalServerError,
			StatusText:      "Error:" + err.Error(),
			Service:         nil,
			ServiceSeverity: nil,
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	//	service table check
	s = `SELECT COALESCE(COUNT(x.SERVICE_NAME),0) AS cc FROM lms.service_logs x
	     WHERE (x.service_name = ?) AND (x.severity = ?);`

	SeverityCount := 0
	err = repo.App.Conn.DB.QueryRowContext(ctx, s, servName, severity).Scan(&SeverityCount)
	if err != nil {
		resp := Result{
			Status:     http.StatusInternalServerError,
			StatusText: "Error:" + err.Error(),
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if servCount == 0 && SeverityCount == 0 {
		resp := Result{
			Status:     http.StatusOK,
			StatusText: "Service and Severity not found! OK",
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	} else if servCount == SeverityCount {
		resp := Result{
			Status:     http.StatusOK,
			StatusText: "Service and Severity count match! OK",
			Service: ServiceLogs{
				ServiceName: servName,
				Severity:    severity,
				Count:       servCount,
			},
			ServiceSeverity: ServiceSeverity{
				ServiceName: servName,
				Severity:    severity,
				Count:       SeverityCount,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	} else {
		resp := Result{
			Status:     http.StatusOK,
			StatusText: "Service and Severity not match! NOT OK!",
			Service: ServiceLogs{
				ServiceName: servName,
				Severity:    severity,
				Count:       servCount,
			},
			ServiceSeverity: ServiceSeverity{
				ServiceName: servName,
				Severity:    severity,
				Count:       SeverityCount,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
}
