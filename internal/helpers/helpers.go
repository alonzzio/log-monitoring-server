package helpers

import (
	"bytes"
	"errors"
	"github.com/alonzzio/envr"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// FindSpecificFileNames finds file names without path inside a folder
func FindSpecificFileNames(root, pattern string) ([]string, error) {
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

// LoadEnv loads env file from directory. Add env file containing folder and file name
func LoadEnv(envDirectory string, filenames ...string) error {
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

	return nil
}

// LoadENVtoConfig loads env variables to App config
func LoadENVtoConfig(app *config.AppConfig) error {
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

	n, err = envr.GetInt("SERVICESNAMEPOOLSIZE")
	if err != nil {
		return err
	}
	app.Environments.PubSub.ServiceNamePool = uint(n)

	n, err = envr.GetInt("SERVICEPUBLISHERS")
	if err != nil {
		return err
	}
	app.Environments.PubSub.ServicePublishers = uint(n)

	n, err = envr.GetInt("MESSAGEPERBATCH")
	if err != nil {
		return err
	}
	app.Environments.PubSub.MessageBatch = uint(n)

	n, err = envr.GetInt("MESSAGEFREQUENCY")
	if err != nil {
		return err
	}
	app.Environments.PubSub.MessageFrequency = uint(n)

	// Data Access Layer
	s, err = envr.GetString("DALPORTNUMBER")
	if err != nil {
		return err
	}
	app.Environments.DataAccessLayer.PortNumber = s

	// Data Collection Layer
	n, err = envr.GetInt("DCLNUMBEROFWORKERS")
	if err != nil {
		return err
	}
	app.Environments.DataCollectionLayer.Workers = uint(n)

	n, err = envr.GetInt("DCLJOBSBUFFER")
	if err != nil {
		return err
	}
	app.Environments.DataCollectionLayer.JobsBuffer = uint(n)

	n, err = envr.GetInt("DCLRESULTBUFFER")
	if err != nil {
		return err
	}
	app.Environments.DataCollectionLayer.ResultBuffer = uint(n)

	n, err = envr.GetInt("DCLRECEIVERGOROUTINES")
	if err != nil {
		return err
	}
	app.Environments.DataCollectionLayer.ReceiverGoRoutines = uint(n)

	n, err = envr.GetInt("DCLRECEIVERGOROUTINES")
	if err != nil {
		return err
	}
	app.Environments.DataCollectionLayer.ReceiverGoRoutines = uint(n)

	n, err = envr.GetInt("DCLRECIEVERTIMEOUT")
	if err != nil {
		return err
	}
	app.Environments.DataCollectionLayer.ReceiverTimeOut = uint(n)

	return nil
}

// GetGoRoutineID return the current goroutine id
func GetGoRoutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	// ignoring error here
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
