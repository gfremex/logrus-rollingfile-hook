package fsrollhook

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/KerwinKoo/logrus"
)

// FsrollHook main rolling file hook struck
// File name pattern, e.g. /tmp/tbrfh/2006/01/02/15/minute.04.log
type FsrollHook struct {
	levels          []logrus.Level   // Log levels allowed
	formatter       logrus.Formatter // Log entry formatter
	FileNamePattern string           //e.g. /tmp/tbrfh/2006/01/02/15/minute.04.log
	ConstantPath    string
	file            *os.File    // Pointer of the file
	timer           *time.Timer // Timer to trigger file rollover
	queue           chan *logrus.Entry
	mu              *sync.Mutex
}

// NewHook Create a new FsrollHook.
func NewHook(levels []logrus.Level, formatter logrus.Formatter, fileNamePattern string) (*FsrollHook, error) {
	hook := &FsrollHook{}

	hook.levels = levels
	hook.formatter = formatter
	hook.FileNamePattern = fileNamePattern
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
func (hook *FsrollHook) rolloverAfter() time.Duration {
	// Get the current local time
	t := time.Now().Local()

	oldFileName := t.Format(hook.FileNamePattern)

	var t1 time.Time
	var newFileName string

	t1 = t.Add(time.Minute)
	newFileName = t1.Format(hook.FileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per minute
		t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), t1.Minute(), 0, 0, t1.Location())
		return t2.Sub(t)
	}

	t1 = t.Add(time.Hour)
	newFileName = t1.Format(hook.FileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per hour

		t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), 0, 0, 0, t1.Location())

		return t2.Sub(t)
	}

	t1 = t.AddDate(0, 0, 1)
	newFileName = t1.Format(hook.FileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per day
		t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())

		return t2.Sub(t)
	}

	t1 = t.AddDate(0, 1, 0)
	newFileName = t1.Format(hook.FileNamePattern)
	if oldFileName != newFileName {
		// Need to rollover per month
		t2 := time.Date(t1.Year(), t1.Month(), 1, 0, 0, 0, 0, t1.Location())

		return t2.Sub(t)
	}

	t1 = t.AddDate(1, 0, 0)
	newFileName = t1.Format(hook.FileNamePattern)
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
func (hook *FsrollHook) rolloverFile() (string, error) {
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
	newFileNameOrig := time.Now().Local().Format(hook.FileNamePattern)

	switch strings.ToLower(filepath.Ext(newFileNameOrig)) {
	case GzipSuffix:
		{
			newFileNameOrig = strings.TrimSuffix(newFileNameOrig, GzipSuffix)
		}
	}

	// Create dirs if needed
	dir := filepath.Dir(newFileNameOrig)

	err := os.MkdirAll(dir, os.ModeDir|0755)

	if err != nil {
		return oldFileName, err
	}

	// Create new file
	newFile, err := os.OpenFile(newFileNameOrig, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)

	if err != nil {
		return oldFileName, err
	}

	// Switch hook.file to newFile
	hook.file = newFile

	return oldFileName, nil
}

// Reset timer and archive old file if needed.
func (hook *FsrollHook) resetTimer() {
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
func (hook *FsrollHook) archiveOldFile(fileName string) {
	if archive, ok := Archivers[strings.ToLower(filepath.Ext(hook.FileNamePattern))]; ok {
		err := archive(fileName)

		if err != nil {
			log.Printf("Error on archiving file [%s]: %v\n", fileName, err)
		}
	}
}

// Write logrus.Entry to file.
func (hook *FsrollHook) write(entry *logrus.Entry) error {
	// Acquire the lock
	hook.mu.Lock()

	defer hook.mu.Unlock()

	if hook.file != nil {
		// Writing allowed
		fileName := hook.file.Name()

		_, err := os.Stat(fileName)
		if err != nil && os.IsNotExist(err) {
			dir := filepath.Dir(fileName)
			err := os.MkdirAll(dir, os.ModeDir|0755)
			if err != nil {
				log.Printf("Error on create Dir: %v\n", err)
				return err
			}

			recreateFile, err := os.OpenFile(fileName,
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)

			if err != nil {
				log.Printf("Error on creating new file: %v\n", err)
				return err
			}
			hook.file = recreateFile
		}

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
// if no mesg in hook.queue it while be blocked here
func (hook *FsrollHook) writeEntry() {
	for entry := range hook.queue {
		// Write logrus.Entry to file.
		err := hook.write(entry)

		if err != nil {
			log.Printf("Error on writing to file: %v\n", err)
			return
		}
	}
}

// Levels get levels
func (hook *FsrollHook) Levels() []logrus.Level {
	return hook.levels
}

// Fire logrus fire
func (hook *FsrollHook) Fire(entry *logrus.Entry) error {
	hook.queue <- entry

	return nil
}

// GetFrontFileName get the  front field of filename
// e.g. filename 06.log.2djiDwOoiNNs will return 06.log
func GetFrontFileName(fileName string) string {
	filenameWithSuffix := path.Base(fileName)
	dotIn := strings.Contains(filenameWithSuffix, ".")

	if dotIn == false {
		return fileName
	}

	// nameFields := strings.Split(filenameWithSuffix, ".")
	nameFieldsCount := strings.Count(filenameWithSuffix, ".")
	if nameFieldsCount < 2 {
		return fileName
	}

	fileSuffix := path.Ext(filenameWithSuffix)
	filenameOnly := strings.TrimSuffix(filenameWithSuffix, fileSuffix)

	dir, _ := path.Split(fileName)
	frontFileName := dir + filenameOnly

	return frontFileName
}
