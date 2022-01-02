package collection

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// CreateDbProcessWorkerPools creates a pool of Receiver Workers
func (repo *Repository) CreateDbProcessWorkerPools(poolSize int, logsBatch <-chan LogsBatch, logsBatchReceive chan<- LogsBatch, wg *sync.WaitGroup) {
	wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		// message size can be controlled through env files
		go repo.MessageDbProcessWorker(logsBatch, logsBatchReceive)
	}
	wg.Wait()
}

// MessageDbProcessWorker gets batch the messages from LogsBatch channel and insert to DB
// 5 retries if error occurred. If still error on insert it will send the LogsBatch back to channel
func (repo *Repository) MessageDbProcessWorker(logsBatchReceive <-chan LogsBatch, logsBatchSend chan<- LogsBatch) {
	for {
		select {
		case lb := <-logsBatchReceive:
			retry := 5
			success := false
			for i := 0; i < retry; i++ {
				//err := repo.BulkInsert(lb.LogMessage, lb.ServiceSeverity)
				err := repo.BulkDbInsert(lb.LogMessage, lb.ServiceSeverity)
				if err != nil {
					fmt.Println(err)
					success = false
					continue
				} else {
					// no error mean try was fine.
					success = true
					break
				}
			}
			if !success {
				//all retries are failed.
				//send batch log to channel back
				logsBatchSend <- lb
				// or write to file or as per business logic
			}
		}
	}
}

// BulkDbInsert inserts batch data
func (repo *Repository) BulkDbInsert(messageRows *[]Message, logSeverityRows *[]ServiceSeverity) error {
	valueStrings1 := make([]string, 0, len(*messageRows))
	valueArgs1 := make([]interface{}, 0, len(*messageRows)*4)
	for _, post := range *messageRows {
		valueStrings1 = append(valueStrings1, "(?, ?, ?, ?)")
		valueArgs1 = append(valueArgs1, post.ServiceName)
		valueArgs1 = append(valueArgs1, post.Payload)
		valueArgs1 = append(valueArgs1, post.Severity)
		valueArgs1 = append(valueArgs1, post.Timestamp)
	}
	stmt1 := fmt.Sprintf("INSERT INTO lms.service_logs (service_name,payload,severity,`timestamp`) VALUES %s",
		strings.Join(valueStrings1, ","))

	valueStrings2 := make([]string, 0, len(*logSeverityRows))
	valueArgs2 := make([]interface{}, 0, len(*logSeverityRows)*3)
	for _, post2 := range *logSeverityRows {
		valueStrings2 = append(valueStrings2, "(?, ?, ?)")
		valueArgs2 = append(valueArgs2, post2.ServiceName)
		valueArgs2 = append(valueArgs2, post2.Severity)
		valueArgs2 = append(valueArgs2, post2.Count)
	}
	stmt2 := fmt.Sprintf("INSERT INTO lms.service_severity (service_name,severity,`count`) VALUES %s",
		strings.Join(valueStrings2, ","))

	// process
	ctx := context.Background()
	tx, err := repo.App.Conn.DB.BeginTx(ctx, nil)
	if err != nil {
		fmt.Println("Err on sql begin:", err)
		return err
	}

	_, execErr := tx.ExecContext(ctx, stmt1, valueArgs1...)
	if execErr != nil {
		_ = tx.Rollback()
		fmt.Println("Err on sql exec:", execErr)
		return execErr
	}

	_, execErr = tx.ExecContext(ctx, stmt2, valueArgs2...)
	if execErr != nil {
		_ = tx.Rollback()
		fmt.Println("Err on sql exec:", execErr)
		return execErr
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("Err on commit:", err)
		return err
	}

	return nil
}
