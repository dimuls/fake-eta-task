package web

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dimuls/fake-eta-task/entity"
	"github.com/facebookarchive/freeport"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type WAPIClientMock struct{}

func (c *WAPIClientMock) Cars(_ *entity.Coordinate, _ int) (
	[]*entity.Coordinate, error) {
	return []*entity.Coordinate{{}, {}, {}}, nil
}

func (c *WAPIClientMock) Predicts(_ *entity.Coordinate,
	_ []*entity.Coordinate) ([]int, error) {
	return []int{0, 1, 2}, nil
}

func TestMain(m *testing.M) {
	logrus.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

var client = &fasthttp.Client{}

func BenchmarkServer(b *testing.B) {
	port, err := freeport.Get()
	if err != nil {
		b.Fatal("failed to get free port: " + err.Error())
	}

	s := NewServer(fmt.Sprintf("127.0.0.1:%d", port), &WAPIClientMock{})
	go func() {
		err := s.Start()
		if err != nil {
			b.Fatal("failed to start web server: " + err.Error())
		}
	}()

	defer func() {
		err := s.Stop()
		if err != nil {
			b.Error("failed to stop web server: " + err.Error())
		}
	}()

	// wait until server will start
	time.Sleep(200 * time.Millisecond)

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	url := fmt.Sprintf(
		"http://127.0.0.1:%d/nearest-car?lat=0.1&lng=0.2", port)

	req.SetRequestURI(url)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err := client.Do(req, res)
		if err != nil {
			b.Fatal("failed to do request: " + err.Error())
		}

		if res.StatusCode() != fasthttp.StatusOK {
			b.Fatalf("not OK status code: %d", res.StatusCode())
		}

		res.Reset()
	}
}

func BenchmarkFastServer(b *testing.B) {
	port, err := freeport.Get()
	if err != nil {
		b.Fatal("failed to get free port: " + err.Error())
	}

	s := NewFastServer(fmt.Sprintf("127.0.0.1:%d", port),
		&WAPIClientMock{})
	go func() {
		err := s.Start()
		if err != nil {
			b.Fatal("failed to start web server: " + err.Error())
		}
	}()

	defer func() {
		err := s.Stop()
		if err != nil {
			b.Error("failed to stop web server: " + err.Error())
		}
	}()

	// wait until server will start
	time.Sleep(200 * time.Millisecond)

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	url := fmt.Sprintf(
		"http://127.0.0.1:%d/nearest-car?lat=0.1&lng=0.2", port)

	req.SetRequestURI(url)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err := client.Do(req, res)
		if err != nil {
			b.Fatal("failed to do request: " + err.Error())
		}

		if res.StatusCode() != fasthttp.StatusOK {
			b.Fatalf("not OK status code: %d", res.StatusCode())
		}

		res.Reset()
	}
}
