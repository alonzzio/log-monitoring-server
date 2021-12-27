package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/alonzzio/envr"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

// run initialise project with necessary configurations
// initialise and connect to database
// Creating tables in database
// Loading env from file to config etc.
func run() error {
	p, err := os.Getwd()
	if err != nil {
		return err
	}

	log.Println("ENV files Loading...")
	parent := filepath.Dir(p)

	var fileNames []string
	fileNames, err = findSpecificFileNames(parent+"/cmd/env", "*.env")
	if err != nil {
		return err
	}
	err = loadEnv(parent+"/cmd/env", fileNames...)
	if err != nil {
		return err
	}
	log.Println("ENV Loaded.")

	c, err := newDBConn()
	if err != nil {
		return err
	}

	app.Conn = c

	err = initialiseDatabase(&app)
	if err != nil {
		return err
	}

	return nil
}

// newConn connects to the database and generates a connection pool
// Connection pooling parameters can be accessed using env variables if wanted to
func newDBConn() (*config.Conn, error) {
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

// findSpecificFileNames finds file names without path inside a folder
func findSpecificFileNames(root, pattern string) ([]string, error) {
	var filenames []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			filenames = append(filenames, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return filenames, nil
}

// loadEnv loads env file from directory. Add env file containing folder and file name
func loadEnv(envDirectory string, filenames ...string) error {
	if len(envDirectory) < 1 {
		return errors.New("environment directory not supplied")
	}

	var f []string
	for _, file := range filenames {
		file = envDirectory + "/" + file // building the directory path
		f = append(f, file)
	}

	//loads environment files from  directory
	err := godotenv.Load(f...)
	if err != nil {
		return err
	}

	err = loadENVtoConfig()
	if err != nil {
		return err
	}

	return nil
}

// LoadENVtoConfig loads env variables to App config
func loadENVtoConfig() error {
	n, err := envr.GetInt("SENTENCECOUNT")
	if err != nil {
		return err
	}
	app.Environments.Paragraph.SentenceCount = n

	n, err = envr.GetInt("WORDCOUNT")
	if err != nil {
		return err
	}
	app.Environments.Paragraph.WordCount = n

	n, err = envr.GetInt("SERVICENAMECHARLEGTH")
	if err != nil {
		return err
	}
	app.Environments.ServiceLog.ServiceNameCharLength = uint(n)

	s, err := envr.GetString("PROJECTID")
	if err != nil {
		return err
	}
	app.Environments.PubSub.ProjectID = s

	s, err = envr.GetString("TOPICID")
	if err != nil {
		return err
	}
	app.Environments.PubSub.TopicID = s

	s, err = envr.GetString("SUBSCRIPTIONID")
	if err != nil {
		return err
	}
	app.Environments.PubSub.SubscriptionID = s

	return nil
}
