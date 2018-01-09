package main

import (
	lrh "github.com/gfremex/logrus-rollingfile-hook"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	// fileNamePattern is a string including date/time layouts as used in time.Time.format().
	// fileNamePattern is used for 3 purposes:
	// 1. Create a file to which log messages are written.
	// 2. Decide how often file rollover happens
	// 3. Archive old file if needed (only gzip supported by now)
	// For example:
	// Assuming that the current local time is "2015-12-31T23:59:01+08:00" and
	// fileNamePattern is "/tmp/tbrfh/2006/01/02/15/minute.04.log"
	// It means that the file /tmp/tbrfh/2015/12/31/23/minute.59.log will be created for writing log messages and
	// file rollover happens every minute.
	// At 00:00, the next file will be /tmp/tbrfh/2016/01/01/00/minute.00.log,
	// At 00:01, the next file will be /tmp/tbrfh/2016/01/01/00/minute.01.log,
	// and so on.
	// If fileNamePattern is "/tmp/tbrfh/2006/01/02/15/minute.04.log.gz", old file will be archived
	// (/tmp/tbrfh/2015/12/31/23/minute.59.log.gz) before the new one (/tmp/tbrfh/2016/01/01/00/minute.00.log)
	// is created.

	// Create a new TimeBasedRollingFileHook
	hook, err := lrh.NewTimeBasedRollingFileHook("tbrfh",
		[]logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
		&logrus.JSONFormatter{},
		"/tmp/tbrfh/2006/01/02/15/minute.04.log")

	if err != nil {
		panic(err)
	}

	// Create a new logrus.Logger
	logger := logrus.New()

	// Add hook to logger
	logger.Hooks.Add(hook)

	// Send message to logger
	logger.Debugf("This must not be logged")

	logger.Info("This is an Info msg")

	logger.Warn("This is a Warn msg")

	logger.Error("This is an Error msg")

	// Ensure log messages were written to file
	time.Sleep(time.Second)
}
