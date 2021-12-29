package config

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"time"
)

type AppConfig struct {
	Environments     Environments
	Conn             *Conn
	GrpcPubSubServer PubSubServer
}

type Environments struct {
	Paragraph  Paragraph
	ServiceLog ServiceLog
	PubSub     PubSub
}

// PubSubServer Holds Client Connection for pstest
type PubSubServer struct {
	Conn *grpc.ClientConn
}

// Conn holds the database connection Pool
type Conn struct {
	DB *sql.DB
}

// MyPool holds the Connection pool settings values
type MyPool struct {
	MaxOpenDBConn      int
	MaxIdleDbConn      int
	MaxDbLifeTime      int
	PingContextTimeout int
}

// ConnectSQL creates database pool for MySql
func ConnectSQL(c context.Context, dsn string, pool *MyPool) (*sql.DB, error) {
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
