package db

import (
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"xorm.io/xorm"
)

func pingWithRetry(ping func() error, intervalMs, timeoutMs int) error {
	if intervalMs <= 0 {
		return ping()
	}

	if timeoutMs <= 0 {
		timeoutMs = 60000
	}

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	var err error

	for attempt := 1; ; attempt++ {
		err = ping()
		if err == nil {
			if attempt > 1 {
				log.Infof("Connected to database after %d attempts", attempt)
			}
			return nil
		}

		if time.Now().After(deadline) {
			return err
		}

		log.Warningf("Could not connect to db (attempt %d), retrying in %dms: %v", attempt, intervalMs, err)
		time.Sleep(time.Duration(intervalMs) * time.Millisecond)
	}
}

func pingEngineWithRetry(engine *xorm.Engine) error {
	return pingWithRetry(
		engine.Ping,
		config.DatabaseConnectRetryInterval.GetInt(),
		config.DatabaseConnectRetryTimeout.GetInt(),
	)
}
