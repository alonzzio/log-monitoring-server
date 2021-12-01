package dockertest_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
	"time"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/mysql", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	code := m.Run()
	time.Sleep(time.Second * 120)

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

//goland:noinspection ALL
func TestSomething(t *testing.T) {
	createTables := `
CREATE TABLE service_logs (
	service_name VARCHAR(100) NOT NULL,
	payload VARCHAR(2048) NOT NULL,
	severity ENUM("debug", "info", "warn", "error", "fatal") NOT NULL,
	timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
)`

	createTables2 := `
CREATE TABLE service_severity (
	service_name VARCHAR(100) NOT NULL,
	severity ENUM("debug", "info", "warn", "error", "fatal") NOT NULL,
	count INT(4) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL)`

	_, err := db.Exec(createTables)
	if err != nil {
		//TODO Handle error
		log.Fatalln(err)
	}

	_, err = db.Exec(createTables2)
	if err != nil {
		//TODO Handle error
		log.Fatalln(err)
	}
	t.Log("created")
	t.Logf("createdd")

}

type Message struct {
	ServiceName string    `json:"service_name"`
	Payload     string    `json:"payload"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}
