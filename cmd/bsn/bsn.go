package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"

	bootstrap2 "github.com/digiu-ai/p2p-bridge/node/bootstrap"
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
	var listen string
	var port uint
	var name string
	var logLevel int
	var pprofFlag bool
	var keysPath string
	flag.StringVar(&mode, "mode", "start", "run \"./bridge -mode init\" to init node")
	flag.StringVar(&listen, "listen", "0.0.0.0", "listen ip address")
	flag.UintVar(&port, "port", 4001, "-port")
	flag.StringVar(&name, "name", "bootstrap", "name for key and config files")
	flag.IntVar(&logLevel, "verbosity", int(logrus.InfoLevel), "run -verbosity 6 to set Trace loglevel")
	flag.BoolVar(&pprofFlag, "profiling", false, "run with '-profiling true' argument to use profiler on \"http://localhost:1234/debug/pprof/\"")
	flag.StringVar(&keysPath, "keys-path", "keys", "keys directory path")
	flag.Parse()

	if pprofFlag == true {
		initPprof()
	}

	logrus.SetLevel(logrus.Level(logLevel))

	// Toggle logrus log level between current logLevel and trace by USR2 os signal
	runa.LogrusLevelHandler(logrus.Level(logLevel))

	keysPath = strings.TrimSuffix(keysPath, "/")
	logrus.Tracef("mode: %s, name: %s, keys-path: %s", mode, name, keysPath)

	switch mode {
	case "init":
		err := bootstrap2.NodeInit(keysPath, name, listen, port)
		if err != nil {
			logrus.Error(fmt.Errorf("node init error %w", err))
		}
	case "start":
		err := bootstrap2.NewNode(keysPath, name, listen, port)
		if err != nil {
			logrus.Error(fmt.Errorf("node start error %w", err))
		}
	default:
		logrus.Fatalf("invalid mode: %s", mode)
	}
	if mode == "init" {

	} else {

	}

}
