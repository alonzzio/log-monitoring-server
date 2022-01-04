package logging

import (
	"github.com/rs/zerolog"

	"os"
)

type Logging struct {
	*zerolog.Logger
}

func (l Logging) NewStdOut() *Logging {
	lg := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &Logging{&lg}
}

func (l Logging) NewFileLog(fName string) (*Logging, error) {
	f, err := os.OpenFile(fName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	lg := zerolog.New(f).With().Timestamp().Logger()
	return &Logging{&lg}, nil
}
