package log

import (
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"github.com/op/go-logging"
	"xorm.io/xorm/log"
)

type MailLogger struct {
	logger *logging.Logger
	level  log.LogLevel
}

const mailFormat = `%{color}%{time:` + time.RFC3339Nano + `}: %{level}` + "\t" + `▶ [MAIL] %{id:03x}%{color:reset} %{message}`
const mailLogModule = `vikunja_mail`

func NewMailLogger() *MailLogger {
	lvl := strings.ToUpper(config.LogMailLevel.GetString())
	level, err := logging.LogLevel(lvl)
	if err != nil {
		Criticalf("Error setting database log level: %s", err.Error())
	}

	mailLogger := &MailLogger{
		logger: logging.MustGetLogger(mailLogModule),
	}

	logBackend := logging.NewLogBackend(GetLogWriter("mail"), "", 0)
	backend := logging.NewBackendFormatter(logBackend, logging.MustStringFormatter(mailFormat+"\n"))

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(level, mailLogModule)

	mailLogger.logger.SetBackend(backendLeveled)

	switch level {
	case logging.CRITICAL:
	case logging.ERROR:
		mailLogger.level = log.LOG_ERR
	case logging.WARNING:
		mailLogger.level = log.LOG_WARNING
	case logging.NOTICE:
	case logging.INFO:
		mailLogger.level = log.LOG_INFO
	case logging.DEBUG:
		mailLogger.level = log.LOG_DEBUG
	default:
		mailLogger.level = log.LOG_OFF
	}

	return mailLogger
}

func (m *MailLogger) Errorf(format string, v ...interface{}) {
	m.logger.Errorf(format, v...)
}

func (m *MailLogger) Warnf(format string, v ...interface{}) {
	m.logger.Warningf(format, v...)
}

func (m *MailLogger) Infof(format string, v ...interface{}) {
	m.logger.Infof(format, v...)
}

func (m *MailLogger) Debugf(format string, v ...interface{}) {
	m.logger.Debugf(format, v...)
}
