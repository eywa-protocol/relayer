package sentry

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/evalphobia/logrus_sentry"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
)

var (
	hook *logrus_sentry.SentryHook
	mx   sync.Mutex
)

func Init(appName string) {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return
	}

	mx.Lock()
	defer mx.Unlock()
	if hook != nil {
		return
	}

	if appName == "" {
		appName = os.Args[0]
	}

	version := common.Version
	if version == "" {
		version = "v0.0.0"
	}

	tags := map[string]string{
		"app":     appName,
		"version": common.Version,
	}

	if common.Commit != "" {
		tags["commit"] = common.Commit
	}
	if common.BuildTime != "" {
		tags["build-time"] = common.BuildTime
	}

	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}

	if hostName := os.Getenv("EYWA_HOSTNAME"); hostName != "" {
		tags["eywa_hostname"] = hostName
	}

	var (
		err error
	)
	if hook, err = logrus_sentry.NewWithTagsSentryHook(dsn, tags, levels); err != nil {
		logrus.Fatal(fmt.Errorf("can not init sentry on error: %w", err))
	}

	if hookTimeout, err := time.ParseDuration(os.Getenv("SENTRY_TIMEOUT")); err != nil {
		logrus.Warnf("invalid SENTRY_TIMEOUT environment value set to default 10s")
		hook.Timeout = 10 * time.Second
	} else {
		hook.Timeout = hookTimeout
	}

	hook.StacktraceConfiguration.Enable = true

	logrus.AddHook(hook)
}

func AddTags(tags map[string]string) {
	mx.Lock()
	defer mx.Unlock()
	if hook != nil {
		hook.SetTagsContext(tags)
	}
}
