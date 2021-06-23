package run

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/sirupsen/logrus"
)

func Host(h host.Host, cancel func()) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logrus.Infof("\rExiting...\n")

	cancel()

	if err := h.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}
