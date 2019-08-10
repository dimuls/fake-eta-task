package web

import (
	"encoding/json"
	"time"

	"github.com/dimuls/fake-eta-task/entity"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type FastHandler struct {
	wapi WAPIClient
	log  *logrus.Entry
}

func (h *FastHandler) Handle(ctx *fasthttp.RequestCtx) {
	st := time.Now()

	defer func() {
		h.log.WithFields(logrus.Fields{
			"duration_ms": time.Now().Sub(st).Nanoseconds() / msInNs,
			"path":        string(ctx.URI().Path()),
			"query":       string(ctx.URI().QueryString()),
		}).Info("request handled")
	}()

	switch string(ctx.Method()) {
	case fasthttp.MethodGet:
		switch string(ctx.URI().Path()) {
		case "/nearest-car":
			h.NearestCar(ctx)
		default:
			ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
		}
	default:
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
	}
}

func (h *FastHandler) NearestCar(ctx *fasthttp.RequestCtx) {

	args := ctx.QueryArgs()

	lat, err := args.GetUfloat("lat")
	if err != nil {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	lng, err := args.GetUfloat("lat")
	if err != nil {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	targetCoord := &entity.Coordinate{
		Latitude:  lat,
		Longitude: lng,
	}

	// TODO: move limit to config
	carsCoord, err := h.wapi.Cars(targetCoord, 10)
	if err != nil {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	if len(carsCoord) == 0 {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	predicts, err := h.wapi.Predicts(targetCoord, carsCoord)
	if err != nil {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	if len(predicts) != len(carsCoord) {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	min := predicts[0]

	for i := 1; i < len(predicts); i++ {
		if predicts[i] < min {
			min = predicts[i]
		}
	}

	err = json.NewEncoder(ctx.Response.BodyWriter()).Encode(min)
	if err != nil {
		h.log.WithError(err).Error("failed to encode success response")
	}
}
