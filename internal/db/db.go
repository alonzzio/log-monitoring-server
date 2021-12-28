package db

import (
	"context"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"log"
	"os"
	"sync"
)

func InitialiseDatabase(app *config.AppConfig) error {
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

	go ExecuteSQLWorker(sql1, app, errChan, &wg)
	go ExecuteSQLWorker(sql2, app, errChan, &wg)

	wg.Wait()
	close(errChan)

	for err = range errChan {
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

// ExecuteSQLWorker this function executes against DB and passing errors through error channel
// this is not really needed but i am demonstrating the sql can be run parallel
func ExecuteSQLWorker(sql string, app *config.AppConfig, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	// ignoring the result part here
	_, err := app.Conn.DB.Exec(sql)
	if err != nil {
		errChan <- err
	}

	errChan <- nil
}

// NewConn connects to the database and generates a connection pool
// Connection pooling parameters can be accessed using env variables if wanted to
func NewConn() (*config.Conn, error) {
	// Load Mysql Conn Pool
	// docker compose will create lms database
	dsn := fmt.Sprintf("root:%v@tcp(localhost:8084)/%v", os.Getenv("MYSQLROOTPASS"), os.Getenv("MYSQLDBNAME"))
	dbPool := config.MyPool{
		MaxOpenDBConn:      10,
		MaxIdleDbConn:      5,
		MaxDbLifeTime:      300,
		PingContextTimeout: 10,
	}

	ctx := context.Background()
	db, err := config.ConnectSQL(ctx, dsn, &dbPool)
	if err != nil {
		return nil, err
	}

	var conn = &config.Conn{
		DB: db,
	}

	return conn, nil
}
