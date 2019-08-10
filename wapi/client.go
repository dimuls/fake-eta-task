package wapi

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dimuls/fake-eta-task/entity"
	"github.com/sirupsen/logrus"
)

type Cache interface {
	Cars(coord *entity.Coordinate, limit int) (
		cars []*entity.Coordinate, found bool, err error)
	Predicts(target *entity.Coordinate, source []*entity.Coordinate) (
		minutes []int, found bool, err error)
}

type URLsProvider interface {
	URLs() ([]string, error)
}

const (
	carsPath     = "/cars"
	predictsPath = "/predict"
)

type Client struct {
	reqID uint64

	urls  URLsProvider
	cache Cache

	log *logrus.Entry
}

func NewClient(urls URLsProvider, cache Cache) *Client {
	return &Client{
		urls:  urls,
		cache: cache,
		log:   logrus.WithField("subsystem", "wapi_client"),
	}
}

func (c *Client) Cars(coord *entity.Coordinate, limit int) (
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

func (c *Client) cars(baseURL string, coord *entity.Coordinate, limit int) (
	[]*entity.Coordinate, error) {

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.New("failed to parse service URL: " + err.Error())
	}

	u.Path = path.Join(u.Path, carsPath)

	q := u.Query()
	q.Set("lat", strconv.FormatFloat(
		coord.Latitude, 'g', -1, 64))
	q.Set("lng", strconv.FormatFloat(
		coord.Longitude, 'g', -1, 64))
	q.Set("limit", strconv.Itoa(limit))

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.New("failed to create request: " + err.Error())
	}

	// TODO: move timeout to config
	res, err := (&http.Client{Timeout: 1000 * time.Millisecond}).Do(req)
	if err != nil {
		return nil, errors.New("failed to do request: " + err.Error())
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			c.log.WithError(err).Error("failed to close response body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.WithError(err).Error("failed to read all body")
			logrus.WithFields(logrus.Fields{
				"status_code": res.StatusCode,
			}).Error("got not OK status code")
		} else {
			logrus.WithFields(logrus.Fields{
				"status_code": res.StatusCode,
				"body":        string(body),
			}).Error("got not OK status code")
		}
		return nil, errors.New("service returned not OK status code")
	}

	var cars []*entity.Coordinate

	err = json.NewDecoder(res.Body).Decode(&cars)
	if err != nil {
		return nil, errors.New("failed to decode response body: " + err.Error())
	}

	return cars, nil
}

func (c *Client) Predicts(target *entity.Coordinate,
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

func (c *Client) predicts(baseURL string, target *entity.Coordinate,
	sources []*entity.Coordinate) ([]int, error) {

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.New("failed to parse service URL: " + err.Error())
	}

	u.Path = path.Join(u.Path, predictsPath)

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, errors.New("failed to create request: " + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	// TODO: move timeout to config
	res, err := (&http.Client{Timeout: 1000 * time.Millisecond}).Do(req)
	if err != nil {
		return nil, errors.New("failed to do request: " + err.Error())
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			c.log.WithError(err).Error("failed to close response body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.WithError(err).Error("failed to read all body")
			logrus.WithFields(logrus.Fields{
				"status_code": res.StatusCode,
			}).Error("got not OK status code")
		} else {
			logrus.WithFields(logrus.Fields{
				"status_code": res.StatusCode,
				"body":        string(body),
			}).Error("got not OK status code")
		}
		return nil, errors.New("service returned not OK status code")
	}

	var predicts []int

	err = json.NewDecoder(res.Body).Decode(&predicts)
	if err != nil {
		return nil, errors.New("failed to decode response body: " + err.Error())
	}

	return predicts, nil
}
