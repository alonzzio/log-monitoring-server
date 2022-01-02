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

// GetSingleServiceSeverity statistics of  a services
// TODO BUG FIX
func (repo *Repository) GetSingleServiceSeverity(w http.ResponseWriter, r *http.Request) {
	var err error

	type Result struct {
		Status     int         `json:"status"`
		StatusText string      `json:"status_text"`
		Response   interface{} `json:"services_severity"`
	}
	type Count struct {
		CountInBatch *int       `json:"count"  db:"count"`
		CreatedAt    *time.Time `json:"created_at"  db:"created_at"`
	}

	type SeverityName struct {
		Severity   *string  `json:"severity"  db:"severity"`
		BatchCount *[]Count `json:"batch_count"  db:"batch_count"`
	}

	type ServiceName struct {
		Name         *string         `json:"name"  db:"service_name"`
		SeverityName *[]SeverityName `json:"severity_name"  db:"severity_name"`
	}

	servName := r.FormValue("service-name")
	fmt.Println(servName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s := `SELECT service_name,severity,count,created_at
			FROM lms.service_severity
			WHERE service_name = ?
			ORDER BY severity;`

	rows, err := repo.App.Conn.DB.QueryContext(ctx, s, servName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	serv := make([]ServiceName, 0)

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
