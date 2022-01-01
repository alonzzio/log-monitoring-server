package collection

import (
	"fmt"
	"strings"
	"sync"
)

// CreateDbProcessWorkerPools creates a pool of Receiver Workers
func (repo *Repository) CreateDbProcessWorkerPools(poolSize int, logs <-chan Logs, wg *sync.WaitGroup) {
	wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		// message size can be controlled through env files
		// go func
		//go repo.MessageDbProcessWorker(repo.App.Environments.DataCollectionLayer.MessageBatchSize, results, logs)
		go repo.MessageDbProcessWorker(logs)
	}
	wg.Wait()
}

// MessageDbProcessWorker gets the messages from results channel and process as batch send to 'Logs' channel
func (repo *Repository) MessageDbProcessWorker(logs <-chan Logs) {
	for {
		select {
		case l := <-logs:
			fmt.Println("Revd: len:", len(*l))
			err := repo.BulkInsert(l)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("insert happened")
		}
	}
}

func (repo *Repository) BulkInsert(unsavedRows *[]Message) error {
	valueStrings := make([]string, 0, len(*unsavedRows))
	valueArgs := make([]interface{}, 0, len(*unsavedRows)*4)
	for _, post := range *unsavedRows {
		valueStrings = append(valueStrings, "(?, ?, ?, ?)")
		valueArgs = append(valueArgs, post.ServiceName)
		valueArgs = append(valueArgs, post.Payload)
		valueArgs = append(valueArgs, post.Severity)
		valueArgs = append(valueArgs, post.Timestamp)
	}
	stmt := fmt.Sprintf("INSERT INTO lms.service_logs (service_name,payload,severity,`timestamp`) VALUES %s",
		strings.Join(valueStrings, ","))
	_, err := repo.App.Conn.DB.Exec(stmt, valueArgs...)
	return err
}
