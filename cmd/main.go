package main

import (
	"flag"
	p "path"
	"path/filepath"
	"strings"

	common2 "github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/DigiU-Lab/p2p-bridge/node"
	"github.com/sirupsen/logrus"
)

func main() {
	var mode string
	var path string
	var port uint
	flag.StringVar(&mode, "mode", "serve", "relayer mode. Default is serve")
	flag.StringVar(&path, "cnf", "bootstrap.env", "config file absolute path")
	flag.UintVar(&port, "port", 0, "ws port")
	flag.Parse()
	logrus.Tracef("mode %v path %v", mode, path)
	file := filepath.Base(path)
	fname := strings.TrimSuffix(file, p.Ext(file))
	logrus.Tracef("FILE", fname)
	err := common2.GenECDSAKey(fname)
	if err != nil {
		panic(err)
	}
	_, _, err = common2.CreateBN256Key(fname)
	if err != nil {
		panic(err)
	}
	if mode == "init" {
		err = node.NodeInit(path, fname)
		if err != nil {
			logrus.Fatalf("nodeInit %v", err)
			panic(err)
		}

	} else if mode == "singlenode" {
		logrus.Info("Enabled single node mode")
		err = node.NewSingleNode(path)
		if err != nil {
			logrus.Fatalf("NewSingleNode %v", err)
			panic(err)
		}

	} else {
		err = node.NewNode(path, fname, int(port))
		if err != nil {
			logrus.Fatalf("NewNode %v", err)
			panic(err)
		}

	}

}
