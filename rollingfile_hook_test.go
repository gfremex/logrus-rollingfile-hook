package logrus_rollingfile_hook

import (
	"log"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

var testPeriod = 200 * time.Millisecond

var testErrChan = make(chan error)

func testTBRFH(hookId, fileNamePattern string, runDuration time.Duration) error {
	// Create a new TimeBasedRollingFileHook
	hook, err := NewTimeBasedRollingFileHook(hookId,
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

		logger.Infof("No. %7d: %9s", i, hookId)

		logger.Warnf("No. %7d: %9s", i, hookId)

		logger.Errorf("No. %7d: %9s", i, hookId)

		time.Sleep(testPeriod)
	}

	return nil
}

func testRolloverPerMin(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per minute
		testErrChan <- testTBRFH("minute", "/tmp/tbrfh/%Y/%m/%d/%H/minute.%M.log", runDuration)
	}()

	go func() {

		// Test with: gzipped, rollover per minute
		testErrChan <- testTBRFH("minute_gz", "/tmp/tbrfh/%Y/%m/%d/%H/minute_gz.%M.log.gz", runDuration)

	}()
}

func testRolloverPerHour(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per hour
		testErrChan <- testTBRFH("hour", "/tmp/tbrfh/%Y/%m/%d/hour.%H.log", runDuration)

	}()

	go func() {

		// Test with: gzipped, rollover per hour
		testErrChan <- testTBRFH("hour_gz", "/tmp/tbrfh/%Y/%m/%d/hour_gz.%H.log.gz", runDuration)

	}()
}

func testRolloverPerDay(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per day
		testErrChan <- testTBRFH("day", "/tmp/tbrfh/%Y/%m/day.%d.log", runDuration)

	}()

	go func() {

		// Test with: gzipped, rollover per day
		testErrChan <- testTBRFH("day_gz", "/tmp/tbrfh/%Y/%m/day_gz.%d.log.gz", runDuration)

	}()
}

func testRolloverPerMonth(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per month
		testErrChan <- testTBRFH("month", "/tmp/tbrfh/%Y/month.%m.log", runDuration)

	}()

	go func() {

		// Test with: gzipped, rollover per month
		testErrChan <- testTBRFH("month_gz", "/tmp/tbrfh/%Y/month_gz.%m.log.gz", runDuration)

	}()
}

func testRolloverPerYear(runDuration time.Duration) {
	go func() {

		// Test with: non-gzipped, rollover per year
		testErrChan <- testTBRFH("year", "/tmp/tbrfh/year.%Y.log", runDuration)

	}()

	go func() {

		// Test with: gzipped, rollover per year
		testErrChan <- testTBRFH("year_gz", "/tmp/tbrfh/year_gz.%Y.log.gz", runDuration)

	}()
}

func TestTimeBasedRollingFileHook(t *testing.T) {
	var runDuration = 5 * 60 * 1000 * time.Millisecond

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
