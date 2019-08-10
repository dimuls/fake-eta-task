package wapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/dimuls/fake-eta-task/entity"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type FastClient struct {
	reqID uint64

	http *fasthttp.Client

	urls  URLsProvider
	cache Cache

	log *logrus.Entry
}

func NewFastClient(urls URLsProvider, cache Cache) *FastClient {
	return &FastClient{
		http: &fasthttp.Client{
			ReadTimeout: 1 * time.Second,
		},
		urls:  urls,
		cache: cache,
		log:   logrus.WithField("subsystem", "wapi_client"),
	}
}

func (c *FastClient) Cars(coord *entity.Coordinate, limit int) (
	[]*entity.Coordinate, error) {

	cars, found, err := c.cache.Cars(coord, limit)
	if err == nil {
		if found {
			return cars, nil
		}
	} else {
		c.log.WithError(err).WithFields(logrus.Fields{
			"coordinate": coord,
			"limit":      limit,
		}).Error("failed to get cars from cache")
	}

	urls, err := c.urls.URLs()
	if err != nil {
		return nil, errors.New("failed to get urls: " + err.Error())
	}

	reqID := atomic.AddUint64(&c.reqID, 1) % uint64(len(urls))
	try := 0

	for try < len(urls) {

		baseURL := urls[reqID%uint64(len(urls))]

		cars, err := c.cars(baseURL, coord, limit)
		if err == nil {
			return cars, nil
		}

		c.log.WithError(err).WithFields(logrus.Fields{
			"url": baseURL,
		}).Error("failed to get cars from service")

		reqID++
		try++
	}

	return nil, errors.New("all services are failed")
}

func (c *FastClient) cars(baseURL string, coord *entity.Coordinate, limit int) (
	[]*entity.Coordinate, error) {

	uri := fmt.Sprintf("%s%s?lat=%g&lng=%g&limit=%d", baseURL, carsPath,
		coord.Latitude, coord.Longitude, limit)

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(uri)
	req.Header.Set("Content-Type", "application/json")

	res := fasthttp.AcquireResponse()

	err := c.http.Do(req, res)
	if err != nil {
		return nil, errors.New("failed to do request: " + err.Error())
	}

	body := res.Body()

	if res.StatusCode() != fasthttp.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status_code": res.StatusCode,
			"body":        string(body),
		}).Error("got not OK status code")
		return nil, errors.New("service returned not OK status code")
	}

	cars := make([]*entity.Coordinate, 0, limit)

	err = json.Unmarshal(body, &cars)
	if err != nil {
		return nil, errors.New("failed to unmarshal response body: " + err.Error())
	}

	return cars, nil
}

func (c *FastClient) Predicts(target *entity.Coordinate,
	sources []*entity.Coordinate) ([]int, error) {

	minutes, found, err := c.cache.Predicts(target, sources)
	if err == nil {
		if found {
			return minutes, nil
		}
	} else {
		c.log.WithError(err).WithFields(logrus.Fields{
			"target":  target,
			"sources": sources,
		}).Error("failed to get predicts from cache")
	}

	urls, err := c.urls.URLs()
	if err != nil {
		return nil, errors.New("failed to get urls: " + err.Error())
	}

	reqID := atomic.AddUint64(&c.reqID, 1) % uint64(len(urls))
	try := 0

	for try < len(urls) {

		baseURL := urls[reqID%uint64(len(urls))]

		predicts, err := c.predicts(baseURL, target, sources)
		if err == nil {
			return predicts, nil
		}

		c.log.WithError(err).WithFields(logrus.Fields{
			"url": baseURL,
		}).Error("failed to get predicts from service")

		reqID++
		try++
	}

	return nil, errors.New("all services are failed")
}

func (c *FastClient) predicts(baseURL string, target *entity.Coordinate,
	sources []*entity.Coordinate) ([]int, error) {

	uri := baseURL + predictsPath

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(uri)
	req.Header.Set("Content-Type", "application/json")

	res := fasthttp.AcquireResponse()

	err := c.http.Do(req, res)
	if err != nil {
		return nil, errors.New("failed to do request: " + err.Error())
	}

	body := res.Body()

	if res.StatusCode() != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status_code": res.StatusCode,
			"body":        string(body),
		}).Error("got not OK status code")
		return nil, errors.New("service returned not OK status code")
	}

	predicts := make([]int, 0, len(sources))

	err = json.Unmarshal(body, &predicts)
	if err != nil {
		return nil, errors.New("failed to decode response body: " + err.Error())
	}

	return predicts, nil
}
