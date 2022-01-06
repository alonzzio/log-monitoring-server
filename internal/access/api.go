package access

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackskj/carta"
	"net/http"
	"time"
)

// ServerPing pings the server
func (repo *Repository) ServerPing(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Welcome to Data Access Layer")
	return
}

// GetServices statistics of services
func (repo *Repository) GetServices(w http.ResponseWriter, r *http.Request) {
	var err error

	type Result struct {
		Status     int         `json:"status"`
		StatusText string      `json:"status_text"`
		Response   interface{} `json:"services"`
	}

	type Service struct {
		Name string `json:"name"  db:"service_name"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s := `SELECT 
			COALESCE(service_name,'') AS service_name
			FROM lms.service_severity 
			GROUP BY service_name 
			ORDER BY service_name;`
	rows, err := repo.App.Conn.DB.QueryContext(ctx, s)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}
	serv := make([]Service, 0)

	err = carta.Map(rows, &serv)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	resp := Result{
		Status:     200,
		StatusText: "ok",
		Response:   serv,
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
		Service         interface{} `json:"services"`
		ServiceSeverity interface{} `json:"services_severity"`
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

	fmt.Println(servName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Service table check
	s := `SELECT SUM(x.COUNT)AS cc FROM lms.service_severity x
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

	//	Severity table check
	s = `SELECT COUNT(x.SERVICE_NAME)AS cc FROM lms.service_logs x
	     WHERE (x.service_name = ?) AND (x.severity = ?);`

	SeverityCount := 0
	err = repo.App.Conn.DB.QueryRowContext(ctx, s, servName, severity).Scan(&SeverityCount)
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

	if servCount == SeverityCount {
		resp := Result{
			Status:     200,
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
			Status:     200,
			StatusText: "Service count and Severity count not match! NOT OK",
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
