/*
	Log Monitoring System ( LMS )
	This program is indented to demonstrate the functionalities only.
	Not fully focused on complete error handling in place.
	However, I'll try my best to cover error handling in place

	Copy right: Ratheesh Alon Rajan
*/

package main

import (
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"log"
	"os"
	"sync"
)

// app holds application wide configs
var app config.AppConfig

func init() {
	log.Println("Log monitoring Server starting up...")
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	var conn *config.Conn
	conn, err = newConn()
	if err != nil {
		log.Fatal(err)
	}

	// set conn to App
	// When we reach this point, successful mysql/any other Db connection is ready to use
	app.Conn = conn

	err = initialiseDatabase(&app)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("end")
}

func initialiseDatabase(app *config.AppConfig) error {
	_, err := app.Conn.DB.Exec(`CREATE DATABASE IF NOT EXISTS ` + os.Getenv("MYSQLDBNAME") + `;`)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(2)
	errChan := make(chan error, 2)

	sql1 := `CREATE TABLE IF NOT EXISTS service_logs (
			service_name VARCHAR(100) NOT NULL,
			payload VARCHAR(2048) NOT NULL,
			severity ENUM("debug", "info", "warn", "error", "fatal") NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
			);`

	sql2 := `CREATE TABLE IF NOT EXISTS service_severity (
			service_name VARCHAR(100) NOT NULL,
			severity ENUM("debug", "info", "warn", "error", "fatal") NOT NULL,
			count INT(4) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
			);`

	go executeSQLWorker(sql1, app, errChan, &wg)
	go executeSQLWorker(sql2, app, errChan, &wg)

	wg.Wait()
	close(errChan)

	for err = range errChan {
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

// executeSQLWorker this function executes against DB and passing errors through error channel
// this is not really needed but i am demonstrating the sql can be run parallel
func executeSQLWorker(sql string, app *config.AppConfig, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	// ignoring the result part here
	_, err := app.Conn.DB.Exec(sql)
	if err != nil {
		errChan <- err
	}
	// all good
	errChan <- nil
}
