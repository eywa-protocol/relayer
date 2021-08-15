package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	p "path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/bridge"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/runa"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry"
)

const defaultRendezvous = "mygroupofnodes"

func initPprof() {

	go func() {
		_ = http.ListenAndServe(":1234", nil)
	}()
}

func main() {
	var init bool
	var register bool
	var path string
	var port uint
	var logLevel int
	var pprofFlag bool
	var keysPath string
	var commonRendezvous string
	var printVer bool

	flag.BoolVar(&init, "init", false, "run \"./bridge -init\" to init node")
	flag.BoolVar(&register, "register", false, "run \"./bridge -register\" to register node")
	flag.StringVar(&path, "cnf", "bridge.yaml", "config file absolute path")
	flag.UintVar(&port, "port", 0, "-port")
	flag.IntVar(&logLevel, "verbosity", int(logrus.InfoLevel), "run -verbosity 6 to set Trace loglevel")
	flag.BoolVar(&pprofFlag, "profiling", false, "run with '-profiling true' argument to use profiler on \"http://localhost:1234/debug/pprof/\"")
	flag.StringVar(&commonRendezvous, "randevoue", "", "run \"./bridge -randevoue CUSTOMSTRING\" to setup your group of nodes")
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

	sentry.Init("bridge")

	keysPath = strings.TrimSuffix(keysPath, "/")
	logrus.Tracef("init: %v, path: %s, keys-path: %s", init, path, keysPath)

	if err := config.LoadBridgeConfig(path); err != nil {
		logrus.Fatal(err)
	}

	// override config from flags
	if commonRendezvous != "" {
		config.Bridge.Rendezvous = commonRendezvous
	}

	// set default rendezvous on empty
	if config.Bridge.Rendezvous == "" {
		config.Bridge.Rendezvous = defaultRendezvous
	}

	file := filepath.Base(path)
	name := strings.TrimSuffix(file, p.Ext(file))

	if init {
		err := bridge.InitNode(name, keysPath)
		if err != nil {
			logrus.Error(fmt.Errorf("node init error %w", err))
		}
	} else if register {
		err := bridge.RegisterNode(name, keysPath)
		if err != nil {
			logrus.Error(fmt.Errorf("node init error %w", err))
		}
	} else {
		err := bridge.NewNode(name, keysPath, config.Bridge.Rendezvous)
		if err != nil {
			logrus.Fatalf("not registered Node or no keyfile: %v", err)
		}
	}

}
