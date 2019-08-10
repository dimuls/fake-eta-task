package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dimuls/fake-eta-task/entity"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	wapi WAPIClient
	log  *logrus.Entry
}

const msInNs = 10e6

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st := time.Now()

	defer func() {
		h.log.WithFields(logrus.Fields{
			"duration_ms": time.Now().Sub(st).Nanoseconds() / msInNs,
			"path":        r.URL.Path,
			"query":       r.URL.RawQuery,
		}).Info("request handled")
	}()

	switch r.Method {
	case http.MethodGet:
		switch r.URL.Path {
		case "/nearest-car":
			h.NearestCar(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) NearestCar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	lat, err := strconv.ParseFloat(q.Get("lat"), 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(q.Get("lng"), 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	targetCoord := &entity.Coordinate{
		Latitude:  lat,
		Longitude: lng,
	}

	// TODO: move limit to config
	carsCoord, err := h.wapi.Cars(targetCoord, 10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(carsCoord) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	predicts, err := h.wapi.Predicts(targetCoord, carsCoord)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(predicts) != len(carsCoord) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	min := predicts[0]

	for i := 1; i < len(predicts); i++ {
		if predicts[i] < min {
			min = predicts[i]
		}
	}

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(min)
	if err != nil {
		h.log.WithError(err).Error("failed to encode success response")
	}
}
