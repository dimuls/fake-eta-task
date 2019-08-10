package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dimuls/fake-eta-task/wapi"
	"github.com/dimuls/fake-eta-task/web"
	"github.com/sirupsen/logrus"
)

func main() {
	//wapiCl := wapi.NewClient(
	//	wapi.DummyURLsProvider{"https://dev-api.wheely.com/fake-eta"},
	//	wapi.DummyCache{})

	wapiCl := wapi.NewFastClient(
		wapi.DummyURLsProvider{"https://dev-api.wheely.com/fake-eta"},
		wapi.DummyCache{})

	//s := web.NewServer(":80", wapiCl)

	s := web.NewFastServer(":80", wapiCl)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		logrus.Info("starting web server")

		err := s.Start()
		if err != nil {
			logrus.WithError(err).Fatal("failed to start web server")
		}
	}()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs

		logrus.Info("got interrupt signal, stopping")

		err := s.Stop()
		if err != nil {
			logrus.WithError(err).Fatal("failed to stop web server")
		}
	}()

	wg.Wait()
}
