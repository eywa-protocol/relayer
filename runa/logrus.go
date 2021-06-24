package runa

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

// LogrusLevelHandler wait USR2 os signal for toggle logrus log level between logLevel and trace
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
