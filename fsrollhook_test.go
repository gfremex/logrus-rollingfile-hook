package fsrollhook

import (
	"log"
	"testing"
	"time"

	"github.com/KerwinKoo/logrus"
)

var testPeriod = 200 * time.Millisecond

var testErrChan = make(chan error)

// testTBRFH main test func
func testTBRFH(hookID, fileNamePattern string, runDuration time.Duration) error {
	// Create a new FsrollHook
	hook, err := NewHook(
		[]logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
		&logrus.JSONFormatter{},
		fileNamePattern)

	if err != nil {
		return err
	}

	// Create a new logrus.Logger
	logger := logrus.New()

	// Add hook to logger
	logger.Hooks.Add(hook)

	log.Printf("logger: %v", logger)
	log.Printf("logger.Out: %v", logger.Out)
	log.Printf("logger.Formatter: %v", logger.Formatter)
	log.Printf("logger.Hooks: %v", logger.Hooks)
	log.Printf("logger.Level: %v", logger.Level)

	loop := runDuration.Nanoseconds() / testPeriod.Nanoseconds()

	for i := 0; i < int(loop); i++ {
		logger.Debugf("No. %7d: This must not be logged", i)

		logger.Infof("No. %7d: %9s", i, hookID)

		logger.Warnf("No. %7d: %9s", i, hookID)

		logger.Errorf("No. %7d: %9s", i, hookID)

		time.Sleep(testPeriod)
	}

	return nil
}

func testRolloverPerMin(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per minute
		testErrChan <- testTBRFH("minute", "/tmp/tbrfh/2006/01/02/15/minute.04.log", runDuration)
	}()

	go func() {

		// Test with: gzipped, rollover per minute
		testErrChan <- testTBRFH("minute_gz", "/tmp/tbrfh/2006/01/02/15/minute_gz.04.log.gz", runDuration)

	}()
}

func testRolloverPerHour(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per hour
		testErrChan <- testTBRFH("hour", "/tmp/tbrfh/2006/01/02/hour.15.log", runDuration)

	}()

	go func() {
		// Test with: gzipped, rollover per hour
		testErrChan <- testTBRFH("hour_gz", "/tmp/tbrfh/2006/01/02/hour_gz.15.log.gz", runDuration)

	}()
}

func testRolloverPerDay(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per day
		testErrChan <- testTBRFH("day", "/tmp/tbrfh/2006/01/day.02.log", runDuration)

	}()

	go func() {

		// Test with: gzipped, rollover per day
		testErrChan <- testTBRFH("day_gz", "/tmp/tbrfh/2006/01/day_gz.02.log.gz", runDuration)

	}()
}

func testRolloverPerMonth(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per month
		testErrChan <- testTBRFH("month", "/tmp/tbrfh/2006/month.01.log", runDuration)

	}()

	go func() {

		// Test with: gzipped, rollover per month
		testErrChan <- testTBRFH("month_gz", "/tmp/tbrfh/2006/month_gz.01.log.gz", runDuration)

	}()
}

func testRolloverPerYear(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per year
		testErrChan <- testTBRFH("year", "/tmp/tbrfh/year.2006.log", runDuration)

	}()

	go func() {

		// Test with: gzipped, rollover per year
		testErrChan <- testTBRFH("year_gz", "/tmp/tbrfh/year_gz.2006.log.gz", runDuration)

	}()
}

func TestFsrollHook(t *testing.T) {
	var runDuration = 6 * 1000 * time.Millisecond

	testRolloverPerMin(runDuration)

	testRolloverPerHour(runDuration)

	testRolloverPerDay(runDuration)

	testRolloverPerMonth(runDuration)

	testRolloverPerYear(runDuration)

	for i := 0; i < 10; i++ {
		if err := <-testErrChan; err != nil {
			t.Log(err)
		}
	}
}
