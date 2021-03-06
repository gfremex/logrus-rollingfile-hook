## logrus-rollingfile-hook


A [logrus.Hook](https://godoc.org/github.com/sirupsen/logrus#Hook) which sends log entry to file and supports file rollover and archive by a given file name pattern.

## File name pattern

File name pattern is a string including supported date/time conversion specifications as described in [here](https://github.com/lestrrat-go/strftime#supported-conversion-specifications).

File name pattern is used for 3 purposes:

1. Create a file to which log messages are written
2. Decide how often file rollover happens
3. Archive old file if needed (only gzip supported by now)

For example, assuming that:

- the current local time is "**2015-12-31T23:59:01+08:00**"
- fileNamePattern = "**/tmp/tbrfh/%Y/%m/%d/%H/minute.%M.log**"

It means that the file ***/tmp/tbrfh/2015/12/31/23/minute.59.log*** will be created for writing log messages and
file rollover happens every minute.

At ***00:00***, the next file will be ***/tmp/tbrfh/2016/01/01/00/minute.00.log***,
At ***00:01***, the next file will be ***/tmp/tbrfh/2016/01/01/00/minute.01.log***,
and so on.

If fileNamePattern is "**/tmp/tbrfh/%Y/%m/%d/%H/minute.%M.log.gz**", old file will be archived (***/tmp/tbrfh/2015/12/31/23/minute.59.log.gz***) before the new one (***/tmp/tbrfh/2016/01/01/00/minute.00.log***) is created.

***Note:***

***File name pattern has been changed***. It used to be a string including date/time layouts as used in [Time.format(layout string)](https://golang.org/pkg/time/#Time.Format). The layout string includes numbers with special meanings such as 15 (for hour), etc. It’s impossible to use these special numbers as real numbers to construct filenames. However, package [strftime](https://github.com/lestrrat-go/strftime) provide another better way except for incompatibility with [Time.format(layout string)](https://golang.org/pkg/time/#Time.Format).

For backward compatibility, please checkout [logrus-rollingfile-hook.v1](https://gopkg.in/gfremex/logrus-rollingfile-hook.v1)

## How to use

### Import package

```Go
import lrh "github.com/gfremex/logrus-rollingfile-hook"
```

### Create a hook (TimeBasedRollingFileHook)

```Go
NewTimeBasedRollingFileHook(id string, levels []logrus.Level, formatter logrus.Formatter, fileNamePattern string) (*TimeBasedRollingFileHook, error)
```

- id: Hook Id
- levels: [logrus.Levels](https://godoc.org/github.com/sirupsen/logrus#Level) supported by the hook
- formatter: [logrus.Formatter](https://godoc.org/github.com/sirupsen/logrus#Formatter) used by the hook
- fileNamePattern: File name pattern

For example:

```Go
hook, err := lrh.NewTimeBasedRollingFileHook("tbrfh",
		[]logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
		&logrus.JSONFormatter{},
		"/tmp/tbrfh/%Y/%m/%d/%H/minute.%M.log")
```

### Create a [logrus.Logger](https://godoc.org/github.com/sirupsen/logrus#Logger)

For example:

```Go
logger := logrus.New()
```

### Add hook to logger

```Go
logger.Hooks.Add(hook)
```

### Send messages to logger

For example:

```Go
logger.Debug("This must not be logged")

logger.Info("This is an Info msg")

logger.Warn("This is a Warn msg")

logger.Error("This is an Error msg")
```

#### Complete examples

[https://github.com/gfremex/logrus-rollingfile-hook/tree/master/examples](https://github.com/gfremex/logrus-rollingfile-hook/tree/master/examples)
