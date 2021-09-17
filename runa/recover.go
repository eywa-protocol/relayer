package runa

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func MainRecoveryExit(appName string) {
	if r := recover(); r != nil {
		err := fmt.Errorf("%s exit on panic recovered: %w", appName, r.(error))

		logrus.WithField("debug", string(debug.Stack())).Error(err)
		os.Exit(1)
	}

	logrus.Infof("%s normal exit", appName)
	os.Exit(0)
}
