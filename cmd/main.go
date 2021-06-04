package main

import (
	"flag"
	"github.com/DigiU-Lab/p2p-bridge/node"
	"github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"os"
	p "path"
	"path/filepath"
	"runtime/pprof"
	"strings"
)

func initPprof() {
	f, err := os.Create("go-pprof.log")
	if err != nil {
		logrus.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close() // error handling omitted for example./
	if err := pprof.StartCPUProfile(f); err != nil {
		logrus.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
	go func() {
		http.ListenAndServe(":1234", nil)
	}()
}

func main() {
	initPprof()
	var mode string
	var path string
	var port uint
	var logLevel int
	flag.StringVar(&mode, "mode", "serve", "relayer mode. Default is serve")
	flag.StringVar(&path, "cnf", "bootstrap.env", "config file absolute path")
	flag.UintVar(&port, "port", 0, "ws port")
	flag.IntVar(&logLevel, "v", int(logrus.InfoLevel), "loglevel logrus.Tracelevel")
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
