package config

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type AppConfig struct {
	Environments     Environments
	Conn             *Conn
	GrpcPubSubServer PubSubServer
	Logger           Logger
}

type Environments struct {
	Paragraph           Paragraph
	ServiceLog          ServiceLog
	PubSub              PubSub
	DataAccessLayer     DataAccessLayer
	DataCollectionLayer DataCollectionLayer
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

type Logger struct {
	Logger zerolog.Logger
}
