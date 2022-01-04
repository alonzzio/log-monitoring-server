package collection

import (
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/rs/zerolog"
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
	logg = r.App.Logger.Logger
}

var Repo *Repository

var logg zerolog.Logger
