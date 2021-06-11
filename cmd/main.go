package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	p "path"
	"path/filepath"
	"strings"

	"github.com/digiu-ai/p2p-bridge/node"
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
	flag.StringVar(&mode, "mode", "serve", "run \"./bridge -mode init\" to init node")
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
	keysPath = strings.TrimSuffix(keysPath, "/")
	logrus.Tracef("mode: %s, path: %s, keys-path: %s", mode, path, keysPath)
	file := filepath.Base(path)
	fname := strings.TrimSuffix(file, p.Ext(file))
	if mode == "init" {
		err := node.NodeInit(path, fname, keysPath)
		if err != nil {
			logrus.Errorf("nodeInit %v", err)
		}

	} else {
		err := node.NewNode(path, fname, commonRendezvous)
		if err != nil {
			logrus.Fatalf("not registered Node or no keyfile: %v", err)
		}

	}

}
