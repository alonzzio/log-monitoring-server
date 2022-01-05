package lmslogging

import (
	"github.com/rs/zerolog"

	"os"
)

type LmsLogging struct {
	SysLog *zerolog.Logger
	AppLog *zerolog.Logger
}

type Severity int

const (
	Debug Severity = iota
	Info
	Warn
	Error
	Fatal
)

type Log struct {
	SysLog   bool
	Severity Severity
	Prefix   string
	Message  string
}

// NewSysFileLog initialise and return LmsLogging
func (l LmsLogging) NewSysFileLog(fName string) (*LmsLogging, error) {
	f, err := os.OpenFile(fName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	lg := zerolog.New(f).With().Timestamp().Logger()
	return &LmsLogging{SysLog: &lg}, nil
}

// NewAppFileLog initialise and return LmsLogging
func (l LmsLogging) NewAppFileLog(fName string) (*LmsLogging, error) {
	f, err := os.OpenFile(fName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	appLog := zerolog.New(f).With().Timestamp().Logger()
	return &LmsLogging{AppLog: &appLog}, nil
}

// NewSysAndAppFileLog initialise and return LmsLogging
func (l LmsLogging) NewSysAndAppFileLog(sysLog, AppLog string) (*LmsLogging, *os.File, *os.File, error) {
	f, err := os.OpenFile(sysLog,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, nil, err
	}

	ff, errr := os.OpenFile(AppLog,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errr != nil {
		return nil, nil, nil, err
	}

	sl := zerolog.New(f).With().Timestamp().Logger()
	al := zerolog.New(ff).With().Timestamp().Logger()
	return &LmsLogging{SysLog: &sl, AppLog: &al}, f, ff, nil
}

// LogWriter writes the log to the file
// Log receives from Log channel
func (l *LmsLogging) LogWriter(logs <-chan Log) {
	for {
		select {
		case lgs := <-logs:
			if lgs.SysLog {
				switch lgs.Severity {
				case Debug:
					l.SysLog.Debug().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				case Info:
					l.SysLog.Info().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				case Warn:
					l.SysLog.Warn().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				case Error:
					l.SysLog.Error().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				case Fatal:
					l.SysLog.Fatal().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				default:
				}
			} else {
				switch lgs.Severity {
				case Debug:
					l.AppLog.Debug().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				case Info:
					l.AppLog.Info().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				case Warn:
					l.AppLog.Warn().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				case Error:
					l.AppLog.Error().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				case Fatal:
					l.AppLog.Fatal().Str("Origin", lgs.Prefix).Msg(lgs.Message)
				default:
				}
			}
		}
	}
}
