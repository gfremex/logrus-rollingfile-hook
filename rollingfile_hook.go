package logrus_rollingfile_hook

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type TimeBasedRollingFileHook struct {
	// Id of the hook
	id string

	// Log levels allowed
	levels []logrus.Level

	// Log entry formatter
	formatter logrus.Formatter

	// File name pattern, e.g. /tmp/tbrfh/2006/01/02/15/minute.04.log
	fileNamePattern string

	// Pointer of the file
	file *os.File

	// Timer to trigger file rollover
	timer *time.Timer

	queue chan *logrus.Entry

	mu *sync.Mutex
}

// Create a new TimeBasedRollingFileHook.
func NewTimeBasedRollingFileHook(id string, levels []logrus.Level, formatter logrus.Formatter, fileNamePattern string) (*TimeBasedRollingFileHook, error) {
	hook := &TimeBasedRollingFileHook{}

	hook.id = id
	hook.levels = levels
	hook.formatter = formatter
	hook.fileNamePattern = fileNamePattern
	hook.queue = make(chan *logrus.Entry, 1000)
	hook.mu = &sync.Mutex{}

	// Create new file
	_, err := hook.rolloverFile()

	if err != nil {
		log.Printf("Error on creating new file: %v\n", err)
	}

	// Calculate duration triggering the next rollover
	d := hook.rolloverAfter()

	if d.Nanoseconds() > 0 {
		hook.timer = time.AfterFunc(d, hook.resetTimer)
	}

	// Write logrus.Entry
	go hook.writeEntry()

	return hook, nil
}

// Calculate duration triggering the next rollover.
// There are 5 degrees of rollover: per minute, per hour, per day, per month, per year.
// This function will test each one from the lowest (per minute) to the highest (per year).
// If 0 or negative returned, it means no more rollovers needed.
func (hook *TimeBasedRollingFileHook) rolloverAfter() time.Duration {
	// Get the current local time
	t := time.Now().Local()

	oldFileName := t.Format(hook.fileNamePattern)

	var t1 time.Time
	var newFileName string

	t1 = t.Add(time.Minute)
	newFileName = t1.Format(hook.fileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per minute

		t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), t1.Minute(), 0, 0, t1.Location())

		return t2.Sub(t)
	}

	t1 = t.Add(time.Hour)
	newFileName = t1.Format(hook.fileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per hour

		t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), 0, 0, 0, t1.Location())

		return t2.Sub(t)
	}

	t1 = t.AddDate(0, 0, 1)
	newFileName = t1.Format(hook.fileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per day

		t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())

		return t2.Sub(t)
	}

	t1 = t.AddDate(0, 1, 0)
	newFileName = t1.Format(hook.fileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per month

		t2 := time.Date(t1.Year(), t1.Month(), 1, 0, 0, 0, 0, t1.Location())

		return t2.Sub(t)
	}

	t1 = t.AddDate(1, 0, 0)
	newFileName = t1.Format(hook.fileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per year

		t2 := time.Date(t1.Year(), 1, 1, 0, 0, 0, 0, t1.Location())

		return t2.Sub(t)
	}

	return time.Duration(0)
}

// Roll over file.
// Old file name and error will be returned.
// If Old file does not exist, empty string will be returned.
func (hook *TimeBasedRollingFileHook) rolloverFile() (string, error) {
	// Acquire the lock
	hook.mu.Lock()

	defer hook.mu.Unlock()

	oldFile := hook.file

	// Forbid output to the hook
	hook.file = nil

	var oldFileName string

	// Close old file if needed
	if oldFile != nil {
		oldFileName = oldFile.Name()

		if err := oldFile.Close(); err != nil {
			log.Printf("Error on closing old file [%s]: %v\n", oldFileName, err)
		}
	}

	// Get new file name
	newFileName := time.Now().Local().Format(hook.fileNamePattern)

	switch strings.ToLower(filepath.Ext(newFileName)) {
	case GzipSuffix:
		{
			newFileName = strings.TrimSuffix(newFileName, GzipSuffix)
		}
	}

	// Create dirs if needed
	dir := filepath.Dir(newFileName)

	err := os.MkdirAll(dir, os.ModeDir|0755)

	if err != nil {
		return oldFileName, err
	}

	// Create new file
	newFile, err := os.OpenFile(newFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)

	if err != nil {
		return oldFileName, err
	}

	// Switch hook.file to newFile
	hook.file = newFile

	return oldFileName, nil
}

// Reset timer and archive old file if needed.
func (hook *TimeBasedRollingFileHook) resetTimer() {
	// Roll over file
	oldFileName, err := hook.rolloverFile()

	if err != nil {
		log.Printf("Error on creating new file: %v\n", err)
	}

	// Calculate duration triggering the next rollover
	d := hook.rolloverAfter()

	if d.Nanoseconds() > 0 {
		// Reset timer
		hook.timer.Reset(d)
	} else {
		// No more rollovers needed
		hook.timer.Stop()
	}

	// Archive old file if needed
	if oldFileName != "" {
		go hook.archiveOldFile(oldFileName)
	}
}

// Archive old file if needed.
func (hook *TimeBasedRollingFileHook) archiveOldFile(fileName string) {
	if archive, ok := Archivers[strings.ToLower(filepath.Ext(hook.fileNamePattern))]; ok {
		err := archive(fileName)

		if err != nil {
			log.Printf("Error on archiving file [%s]: %v\n", fileName, err)
		}
	}
}

// Write logrus.Entry to file.
func (hook *TimeBasedRollingFileHook) write(entry *logrus.Entry) error {
	// Acquire the lock
	hook.mu.Lock()

	defer hook.mu.Unlock()

	if hook.file != nil {
		// Writing allowed

		// Format before writing
		b, err := hook.formatter.Format(entry)

		if err != nil {
			return err
		}

		// Writing to file
		_, err = hook.file.Write(b)

		if err != nil {
			return err
		}
	}

	return nil
}

// Write logrus.Entry.
func (hook *TimeBasedRollingFileHook) writeEntry() {
	for entry := range hook.queue {
		// Write logrus.Entry to file.
		err := hook.write(entry)

		if err != nil {
			log.Printf("Error on writing to file: %v\n", err)
		}
	}
}

func (hook *TimeBasedRollingFileHook) Id() string {
	return hook.id
}

func (hook *TimeBasedRollingFileHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *TimeBasedRollingFileHook) Fire(entry *logrus.Entry) error {
	hook.queue <- entry

	return nil
}
