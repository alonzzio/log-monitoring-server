package access

import (
	"github.com/alonzzio/log-monitoring-server/internal/config"
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
