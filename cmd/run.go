package main

import (
	"github.com/alonzzio/log-monitoring-server/internal/db"
	"github.com/alonzzio/log-monitoring-server/internal/helpers"
	"github.com/alonzzio/log-monitoring-server/internal/lmslogging"
	"os"
	"path/filepath"
)

// run initialise project with necessary configurations
// initialise and connect to database
// Creating tables in database
// Loading env from file to config etc.
func run(sysLogger chan<- lmslogging.Log) error {
	p, err := os.Getwd()
	if err != nil {
		return err
	}
	parent := filepath.Dir(p)

	sysLogger <- lmslogging.Log{
		SysLog:   true,
		Severity: lmslogging.Info,
		Prefix:   "AppInitRun",
		Message:  "ENV variables loading",
	}

	var fileNames []string
	fileNames, err = helpers.FindSpecificFileNames(parent+"/cmd/env", "*.env")
	if err != nil {
		return err
	}
	err = helpers.LoadEnv(parent+"/cmd/env", fileNames...)
	if err != nil {
		return err
	}
	sysLogger <- lmslogging.Log{
		SysLog:   true,
		Severity: lmslogging.Info,
		Prefix:   "AppInitRun",
		Message:  "ENV variables loaded.",
	}

	err = helpers.LoadENVtoConfig(&app)
	if err != nil {
		return err
	}

	c, err := db.NewConn()
	if err != nil {
		return err
	}

	app.Conn = c

	err = db.InitialiseDatabase(&app)
	if err != nil {
		return err
	}
	return nil
}
