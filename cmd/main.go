package main

import (
	"flag"
	p "path"
	"path/filepath"
	"strings"

	"github.com/DigiU-Lab/p2p-bridge/node"
	"github.com/sirupsen/logrus"
)

func main() {
	var mode string
	var path string
	var port uint
	var logLevel int
	flag.StringVar(&mode, "mode", "serve", "relayer mode. Default is serve")
	flag.StringVar(&path, "cnf", "bootstrap.env", "config file absolute path")
	flag.UintVar(&port, "port", 0, "ws port")
	flag.IntVar(&logLevel,"v", int(logrus.InfoLevel), "loglevel logrus.Tracelevel" )
	flag.Parse()
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
