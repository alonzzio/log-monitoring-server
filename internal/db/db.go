package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"log"
	"os"
	"sync"
	"time"
)

func InitialiseDatabase(app *config.AppConfig) error {
	_, err := app.Conn.DB.Exec(`CREATE DATABASE IF NOT EXISTS ` + os.Getenv("MYSQLDBNAME") + `;`)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

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

	wg.Add(2)

	go ExecuteSQLWorker(sql1, app, errChan, &wg)
	go ExecuteSQLWorker(sql2, app, errChan, &wg)

	wg.Wait()
	close(errChan)

	for err = range errChan {
		if err != nil {
			log.Fatal(err)
		}
	}

	// After creating tables , truncating tables just make sure this test works fine
	sql1 = `TRUNCATE TABLE lms.service_logs;`
	sql2 = `TRUNCATE TABLE lms.service_severity;`

	errChan2 := make(chan error, 2)
	wg.Add(2)
	// Truncate tables if data exists
	go ExecuteSQLWorker(sql1, app, errChan2, &wg)
	go ExecuteSQLWorker(sql2, app, errChan2, &wg)

	wg.Wait()
	close(errChan2)

	for err = range errChan2 {
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

// ExecuteSQLWorker this function executes against DB and passing errors through error channel
// this is not really needed but, I am demonstrating the sql can be run parallel
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
		MaxOpenDBConn:      5,
		MaxIdleDbConn:      2,
		MaxDbLifeTime:      300,
		PingContextTimeout: 10,
	}

	ctx := context.Background()
	db, err := ConnectSQL(ctx, dsn, &dbPool)
	if err != nil {
		return nil, err
	}

	var conn = &config.Conn{
		DB: db,
	}

	return conn, nil
}

// ConnectSQL creates database pool for MySql
func ConnectSQL(c context.Context, dsn string, pool *config.MyPool) (*sql.DB, error) {
	d, err := newDatabase(dsn)
	if err != nil {
		return nil, err
	}

	d.SetMaxOpenConns(pool.MaxOpenDBConn)
	d.SetMaxIdleConns(pool.MaxIdleDbConn)
	d.SetConnMaxLifetime(time.Duration(pool.MaxDbLifeTime) * time.Minute)

	ctx, cancel := context.WithTimeout(c, time.Duration(pool.PingContextTimeout)*time.Millisecond)
	defer cancel() // releases resources if slowOperation completes before timeout elapses

	if err = d.PingContext(ctx); err != nil {
		return nil, err
	}

	return d, nil
}

// NewDatabase creates new database for the application
func newDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
