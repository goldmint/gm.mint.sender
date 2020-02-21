package http

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	gohttp "net/http"
	"time"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/sender/api/model"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	pkg "github.com/void616/gm.mint.sender/pkg/sender/http"
	"github.com/void616/gm.mint/amount"
)

// send is POST method to request token sending
func (h *HTTP) send(w gohttp.ResponseWriter, r *gohttp.Request) {
	defer r.Body.Close()

	// metrics
	if h.metrics != nil {
		defer func(t time.Time) {
			h.metrics.RequestDuration.WithLabelValues("send").Observe(time.Since(t).Seconds())
		}(time.Now())
	}

	// parse
	req := pkg.SendRequest{}
	{
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			h.logger.WithError(err).Error("Failed to read request")
			return
		}
		if err := json.Unmarshal(b, &req); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal request")
			return
		}
	}

	h.logger.WithField("data", req.String()).Debug("Got sending request")

	// reply
	var res = struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
		Status  int    `json:"-"`
	}{false, "", gohttp.StatusBadRequest}

	defer func() {
		b, err := json.Marshal(&res)
		if err != nil {
			h.logger.WithError(err).Error("Failed to marshal response")
			w.WriteHeader(gohttp.StatusInternalServerError)
			return
		}
		w.WriteHeader(res.Status)
		w.Write(b)
	}()

	// check req service
	if !model.ServiceNameRex.MatchString(req.Service) {
		res.Error = "invalid service name"
		return
	}

	// check req id
	if !model.RequestIDRex.MatchString(req.ID) {
		res.Error = "invalid request ID"
		return
	}

	// parse wallet address
	reqAddr, err := mint.ParsePublicKey(req.PublicKey)
	if err != nil {
		res.Error = "invalid public key"
		return
	}

	// parse token
	reqToken, err := mint.ParseToken(req.Token)
	if err != nil {
		res.Error = "invalid token"
		return
	}

	// valid amount
	reqAmount, err := amount.FromString(req.Amount)
	if err != nil || reqAmount.Value.Cmp(new(big.Int)) <= 0 {
		res.Error = "invalid amount"
		return
	}

	// parse callback
	if req.Callback != "" {
		if !model.ValidCallback(req.Callback) {
			res.Error = "invalid callback"
			return
		}
	}

	// enqueue
	if dups, ok := h.api.EnqueueSending(types.SendingHTTP, req.ID, req.Service, req.Callback, reqAddr, reqAmount, reqToken, req.IgnoreApprovement); !ok {
		if dups {
			res.Error = "request with the same ID registered"
		} else {
			res.Error = "internal failure"
			res.Status = gohttp.StatusInternalServerError
		}
		return
	}

	// success
	res.Success = true
	res.Error = ""
	res.Status = gohttp.StatusOK
}

// approve is POST method to request wallet approvement
func (h *HTTP) approve(w gohttp.ResponseWriter, r *gohttp.Request) {
	defer r.Body.Close()

	// metrics
	if h.metrics != nil {
		defer func(t time.Time) {
			h.metrics.RequestDuration.WithLabelValues("approve").Observe(time.Since(t).Seconds())
		}(time.Now())
	}

	// parse
	req := pkg.ApproveRequest{}
	{
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			h.logger.WithError(err).Error("Failed to read request")
			return
		}
		if err := json.Unmarshal(b, &req); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal request")
			return
		}
	}

	h.logger.WithField("data", req.String()).Debug("Got approvement request")

	// reply
	var res = struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
		Status  int    `json:"-"`
	}{false, "", gohttp.StatusBadRequest}

	defer func() {
		b, err := json.Marshal(&res)
		if err != nil {
			h.logger.WithError(err).Error("Failed to marshal response")
			w.WriteHeader(gohttp.StatusInternalServerError)
			return
		}
		w.WriteHeader(res.Status)
		w.Write(b)
	}()

	// check req service
	if !model.ServiceNameRex.MatchString(req.Service) {
		res.Error = "invalid service name"
		return
	}

	// check req id
	if !model.RequestIDRex.MatchString(req.ID) {
		res.Error = "invalid request ID"
		return
	}

	// parse wallet address
	reqAddr, err := mint.ParsePublicKey(req.PublicKey)
	if err != nil {
		res.Error = "invalid public key"
		return
	}

	// parse callback
	if req.Callback != "" {
		if !model.ValidCallback(req.Callback) {
			res.Error = "invalid callback"
			return
		}
	}

	// enqueue
	if dups, ok := h.api.EnqueueApprovement(types.SendingHTTP, req.ID, req.Service, req.Callback, reqAddr); !ok {
		if dups {
			res.Error = "request with the same ID registered"
		} else {
			res.Error = "internal failure"
			res.Status = gohttp.StatusInternalServerError
		}
		return
	}

	// success
	res.Success = true
	res.Error = ""
	res.Status = gohttp.StatusOK
}
