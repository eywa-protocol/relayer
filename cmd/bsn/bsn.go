package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	bootstrap2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/bootstrap"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/runa"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry"
)

const appName = "bsn"

func initPprof() {

	go func() {
		_ = http.ListenAndServe(":1234", nil)
	}()
}

func main() {
	var init bool
	var listen string
	var port uint
	var name string
	var logLevel int
	var pprofFlag bool
	var keysPath string
	var printVer bool
	flag.BoolVar(&init, "init", false, "run \"./bsn -init\" to init node")
	flag.StringVar(&listen, "listen", "0.0.0.0", "listen ip address")
	flag.UintVar(&port, "port", 4001, "-port")
	flag.StringVar(&name, "name", "bootstrap", "name for key and config files")
	flag.IntVar(&logLevel, "verbosity", int(logrus.InfoLevel), "run -verbosity 6 to set Trace loglevel")
	flag.BoolVar(&pprofFlag, "profiling", false, "run with '-profiling true' argument to use profiler on \"http://localhost:1234/debug/pprof/\"")
	flag.StringVar(&keysPath, "keys-path", "keys", "keys directory path")
	flag.BoolVar(&printVer, "version", false, "print version and exit")
	flag.Parse()

	if printVer {
		common.PrintVersion()
		os.Exit(0)
	}

	if pprofFlag == true {
		initPprof()
	}

	logrus.SetLevel(logrus.Level(logLevel))

	// Toggle logrus log level between current logLevel and trace by USR2 os signal
	runa.LogrusLevelHandler(logrus.Level(logLevel))

	sentry.Init(appName)
	defer runa.MainRecoveryExit(appName)

	keysPath = strings.TrimSuffix(keysPath, "/")
	logrus.Tracef("init: %v, name: %s, keys-path: %s", init, name, keysPath)

	if init {
		err := bootstrap2.NodeInit(keysPath, name, listen, port)
		if err != nil {
			logrus.Error(fmt.Errorf("node init error %w", err))
		}
	} else {
		err := bootstrap2.NewNode(keysPath, name, listen, port)
		if err != nil {
			logrus.Error(fmt.Errorf("node start error %w", err))
		}
	}

}
