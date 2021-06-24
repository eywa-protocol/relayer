package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	p "path"
	"path/filepath"
	"strings"

	"github.com/digiu-ai/p2p-bridge/node/bridge"
	"github.com/digiu-ai/p2p-bridge/runa"
	"github.com/sirupsen/logrus"
)

func initPprof() {

	go func() {
		_ = http.ListenAndServe(":1234", nil)
	}()
}

func main() {
	var mode string
	var path string
	var port uint
	var logLevel int
	var pprofFlag bool
	var keysPath string
	var commonRendezvous string
	flag.StringVar(&mode, "mode", "start", "run \"./bridge -mode init\" to init node")
	flag.StringVar(&path, "cnf", "bootstrap.env", "config file absolute path")
	flag.UintVar(&port, "port", 0, "-port")
	flag.IntVar(&logLevel, "verbosity", int(logrus.InfoLevel), "run -verbosity 6 to set Trace loglevel")
	flag.BoolVar(&pprofFlag, "profiling", false, "run with '-profiling true' argument to use profiler on \"http://localhost:1234/debug/pprof/\"")
	flag.StringVar(&commonRendezvous, "randevoue", "mygroupofnodes", "run \"./bridge -randevoue CUSTOMSTRING\" to setup your group of nodes")
	flag.StringVar(&keysPath, "keys-path", "keys", "keys directory path")
	flag.Parse()

	if pprofFlag == true {
		initPprof()
	}

	logrus.SetLevel(logrus.Level(logLevel))

	// Toggle logrus log level between current logLevel and trace by USR2 os signal
	runa.LogrusLevelHandler(logrus.Level(logLevel))

	keysPath = strings.TrimSuffix(keysPath, "/")
	logrus.Tracef("mode: %s, path: %s, keys-path: %s", mode, path, keysPath)
	file := filepath.Base(path)
	fname := strings.TrimSuffix(file, p.Ext(file))

	switch mode {
	case "init":
		err := bridge.InitNode(path, fname, keysPath)
		if err != nil {
			logrus.Error(fmt.Errorf("node init error %w", err))
		}
	case "start":
		err := bridge.NewNode(path, fname, commonRendezvous)
		if err != nil {
			logrus.Fatalf("not registered Node or no keyfile: %v", err)
		}
	default:
		logrus.Fatalf("invalid mode: %s", mode)
	}
	if mode == "init" {

	} else {

	}

}
