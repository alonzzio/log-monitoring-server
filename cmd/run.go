package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

func run() error {
	p, err := os.Getwd()
	if err != nil {
		return err
	}

	log.Println("ENV files Loading...")
	parent := filepath.Dir(p)

	fmt.Println(parent)
	var fileNames []string
	fileNames, err = findSpecificFileNames(parent+"/cmd/env", "*.env")
	if err != nil {
		return err
	}
	err = loadEnv(parent+"/cmd/env", fileNames...)
	if err != nil {
		return err
	}

	fmt.Println(os.Getenv("MYSQLROOTPASS"))

	return nil
}

// newConn connects to the database and generates a connection pool
// Connection pooling parameters can be accessed using env variables if wanted to
func newConn() (*config.Conn, error) {
	// Load Mysql Conn Pool
	dsn := fmt.Sprintf("root:example@tcp(localhost:8084)/lms")
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

	// loads environment files from  directory
	err := godotenv.Load(f...)
	if err != nil {
		return err
	}

	return nil
}
