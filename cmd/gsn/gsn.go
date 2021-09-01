package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	p "path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/runa"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry"
)

const appName = "gsn"

func initPprof() {

	go func() {
		_ = http.ListenAndServe(":1234", nil)
	}()
}

func main() {
	var (
		path      string
		listen    string
		port      uint
		logLevel  int
		pprofFlag bool
		keysPath  string
		printVer  bool
	)

	flag.StringVar(&path, "cnf", "gsn.yaml", "config file absolute path")
	flag.StringVar(&listen, "listen", "0.0.0.0", "listen ip address")
	flag.UintVar(&port, "port", 4002, "-port")
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
	logrus.Tracef("path: %s, keys-path: %s", path, keysPath)

	if err := config.LoadGsnConfig(path); err != nil {
		logrus.Fatal(err)
	}

	file := filepath.Base(path)
	name := strings.TrimSuffix(file, p.Ext(file))

	if err := gsn.RunNode(name, keysPath, listen, port); err != nil {
		logrus.Fatal(fmt.Errorf("node run error: %w", err))
	}

}
