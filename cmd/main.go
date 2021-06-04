package main

import (
	"flag"
	"github.com/DigiU-Lab/p2p-bridge/node"
	"github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	p "path"
	"path/filepath"
	"strings"
)

func initPprof() {

	go func() {
		http.ListenAndServe(":1234", nil)
	}()
}

func main() {
	var mode string
	var path string
	var port uint
	var logLevel int
	var pprofFlag bool
	flag.StringVar(&mode, "mode", "serve", "run \"./bridge -mode init\" to init node")
	flag.StringVar(&path, "cnf", "bootstrap.env", "config file absolute path")
	flag.UintVar(&port, "port", 0, "-port")
	flag.IntVar(&logLevel, "verbosity", int(logrus.InfoLevel), "run -verbosity 6 to set Trace loglevel")
	flag.BoolVar(&pprofFlag, "profiling", false, "run with '-profiling true' argument to use profiler")
	flag.Parse()
	if pprofFlag == true {
		initPprof()
	}
	logrus.Tracef("mode %v path %v", mode, path)
	file := filepath.Base(path)
	fname := strings.TrimSuffix(file, p.Ext(file))
	logrus.Tracef("FILE", fname)
	logrus.SetLevel(logrus.Level(logLevel))
	if mode == "init" {
		err := node.NodeInit(path, fname)
		if err != nil {
			logrus.Errorf("nodeInit %v", err)
		}

	} else if mode == "singlenode" {
		logrus.Info("Enabled single node mode")
		err := node.NewSingleNode(path)
		if err != nil {
			logrus.Fatalf("NewSingleNode %v", err)
			panic(err)
		}

	} else {
		err := node.NewNode(path, fname)
		if err != nil {
			logrus.Fatalf("not registered Node or no keyfile: %v", err)
		}

	}

}
