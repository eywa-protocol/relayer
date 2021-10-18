package base

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type MetricsServer struct{}

func (s *MetricsServer) Serve(ctx context.Context, listenAddr string, wg *sync.WaitGroup) {

	srv := &http.Server{Addr: listenAddr}
	http.Handle("/metrics", promhttp.Handler())

	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Infof("metrics server listen on %s", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("can not listen and serve metrics server: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctxShutdown); err != nil {
			logrus.Errorf("failed graceful shutdown metrics server on error: %v", err)
		} else {
			logrus.Info("metrics server gracefully shutdown")
		}
	}()

}
