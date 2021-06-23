package run

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

// LogrusLevelHandler Toggle logrus log level between logLevel and trace by USR2 os signal
func LogrusLevelHandler(logLevel logrus.Level) {
	if logLevel != logrus.TraceLevel {
		logLevelChan := make(chan os.Signal, 1)
		defer close(logLevelChan)
		signal.Notify(logLevelChan, syscall.SIGUSR2)
		go func() {
		logLoop:
			for {
				select {
				case sig := <-logLevelChan:
					if sig != nil {
						if logrus.GetLevel() != logrus.TraceLevel {
							logrus.SetLevel(logrus.TraceLevel)
							logrus.Infoln("set loglevel to ", logrus.TraceLevel.String())
						} else {
							logrus.SetLevel(logLevel)
							logrus.Infoln("set loglevel to ", logrus.Level(logLevel).String())
						}
					} else { // signal chan closed
						break logLoop
					}
				}
			}
		}()
	}

}
