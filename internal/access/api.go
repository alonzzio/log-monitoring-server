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

// GetServicesCount statistics of services
func (repo *Repository) GetServicesCount(w http.ResponseWriter, r *http.Request) {
	var err error

	type Result struct {
		Status     int         `json:"status"`
		StatusText string      `json:"status_text"`
		Response   interface{} `json:"service"`
	}

	type Service struct {
		Name             string `json:"name"  db:"service_name"`
		NumberOfServices int    `json:"count"  db:"count"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s := `SELECT 
		COALESCE(service_name,'')AS service_name,
		COUNT(service_name) AS count 
		FROM lms.service_severity 
		GROUP BY service_name;`
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
	json.NewEncoder(w).Encode(resp)

	return
}
