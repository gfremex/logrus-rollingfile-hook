package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KerwinKoo/logrus"

	"log"

	frh "github.com/KerwinKoo/fsrollhook"
)

type MyJSONFormatter struct {
}

// logrus.SetFormatter(new(MyJSONFormatter))

func (f *MyJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry. Consult `godoc` on information about those fields or read the
	// source of the official loggers.
	serialized, err := json.Marshal(entry.Data)
	log.Println("entry is ", entry.Data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

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

	// logFormat := &logrus.JSONFormatter{FieldMap: logrus.FieldMap{
	// 	logrus.FieldKeyMsg:   "msg",
	// 	logrus.FieldKeyLevel: "lvl",
	// 	logrus.FieldKeyTime:  "",
	// }}

	// logFormat.TimestampFormat = "2006-01-02 15:04"

	logFormat := &logrus.TextFormatter{}
	logFormat.DisableColors = true

	hook, err := frh.NewHook(
		[]logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
		logFormat,
		"hhh/2006/01.log")

	if err != nil {
		panic(err)
	}

	// Create a new logrus.Logger
	logger := logrus.New()

	// Add hook to logger
	logger.Hooks.Add(hook)

	// Send message to logger
	logger.Debugf("This must not be logged")

	logger.Info("") //void but be recorded

	logger.Warn("This is a Warn msg")

	logger.Error("This is an Error msg")
	// Ensure log messages were written to file
	time.Sleep(time.Second)
}
