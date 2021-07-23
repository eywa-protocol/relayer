package runa

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/sirupsen/logrus"
)

// Host wait os signals for gracefully shutdown hosts
func Host(h host.Host, cancel func(), wg *sync.WaitGroup) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logrus.Infof("\rExiting...\n")

	cancel()
	wg.Wait()
	if err := h.Close(); err != nil {
		logrus.Error(fmt.Errorf("close host error: %w", err))
	}
	os.Exit(0)
}
