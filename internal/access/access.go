package access

import (
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"net/http"
)

// Repository holds App config
type Repository struct {
	App *config.AppConfig
}

// NewRepo initialise and return Repository Type Which holds AppConfig
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers  sets the repository  for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

var Repo *Repository

// ServerPing pings the server
func (repo *Repository) ServerPing(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Welcome to Data Access Layer")
	return
}
