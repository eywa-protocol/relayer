package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	p "path"
	"path/filepath"
	"strings"

	"github.com/digiu-ai/p2p-bridge/config"
	"github.com/digiu-ai/p2p-bridge/node/bridge"
	"github.com/digiu-ai/p2p-bridge/runa"
	"github.com/sirupsen/logrus"
)

const defaultRendezvous = "mygroupofnodes"

func initPprof() {

	go func() {
		_ = http.ListenAndServe(":1234", nil)
	}()
}

func main() {
	var init bool
	var path string
	var port uint
	var logLevel int
	var pprofFlag bool
	var keysPath string
	var commonRendezvous string
	flag.BoolVar(&init, "init", false, "run \"./bridge -init\" to init node")
	flag.StringVar(&path, "cnf", "bridge.yaml", "config file absolute path")
	flag.UintVar(&port, "port", 0, "-port")
	flag.IntVar(&logLevel, "verbosity", int(logrus.InfoLevel), "run -verbosity 6 to set Trace loglevel")
	flag.BoolVar(&pprofFlag, "profiling", false, "run with '-profiling true' argument to use profiler on \"http://localhost:1234/debug/pprof/\"")
	flag.StringVar(&commonRendezvous, "randevoue", "", "run \"./bridge -randevoue CUSTOMSTRING\" to setup your group of nodes")
	flag.StringVar(&keysPath, "keys-path", "keys", "keys directory path")
	flag.Parse()

	if pprofFlag == true {
		initPprof()
	}

	logrus.SetLevel(logrus.Level(logLevel))

	// Toggle logrus log level between current logLevel and trace by USR2 os signal
	runa.LogrusLevelHandler(logrus.Level(logLevel))

	keysPath = strings.TrimSuffix(keysPath, "/")
	logrus.Tracef("init: %v, path: %s, keys-path: %s", init, path, keysPath)

	if err := config.Load(path); err != nil {
		logrus.Fatal(err)
	}

	// override config from flags
	if commonRendezvous != "" {
		config.App.Rendezvous = commonRendezvous
	}

	// set default rendezvous on empty
	if config.App.Rendezvous == "" {
		config.App.Rendezvous = defaultRendezvous
	}

	file := filepath.Base(path)
	name := strings.TrimSuffix(file, p.Ext(file))

	if init {
		err := bridge.InitNode(name, keysPath)
		if err != nil {
			logrus.Error(fmt.Errorf("node init error %w", err))
		}
	} else {
		err := bridge.NewNode(name, keysPath, config.App.Rendezvous)
		if err != nil {
			logrus.Fatalf("not registered Node or no keyfile: %v", err)
		}
	}

}
