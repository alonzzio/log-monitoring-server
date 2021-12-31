package collection

import (
	"github.com/alonzzio/log-monitoring-server/internal/config"
)

// Repository holds App config
type Repository struct {
	App     *config.AppConfig
	Jobs    chan Job
	Results chan Result
}

// NewRepo initialise and return Repository Type Which holds AppConfig
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App:     a,
		Jobs:    make(chan Job, a.Environments.DataCollectionLayer.JobsBuffer),
		Results: make(chan Result, a.Environments.DataCollectionLayer.ResultBuffer),
	}
}

// NewHandlers  sets the repository  for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

var Repo *Repository
