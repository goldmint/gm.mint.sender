package http

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	gohttp "net/http"
	"net/url"
	"time"
	"unicode/utf8"

	"github.com/void616/gm.mint.sender/internal/sender/api/model"
	mint "github.com/void616/gm.mint"
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
	req := SendRequest{}
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
	if x := utf8.RuneCountInString(req.Service); x < 1 || x > 64 || !model.ServiceNameRex.MatchString(req.Service) {
		res.Error = "Invalid service name"
		return
	}

	// check req id
	if x := utf8.RuneCountInString(req.ID); x < 1 || x > 64 {
		res.Error = "Invalid request ID"
		return
	}

	// parse wallet address
	reqAddr, err := mint.ParsePublicKey(req.PublicKey)
	if err != nil {
		res.Error = "Invalid public key"
		return
	}

	// parse token
	reqToken, err := mint.ParseToken(req.Token)
	if err != nil {
		res.Error = "Invalid token"
		return
	}

	// valid amount
	reqAmount, err := amount.FromString(req.Amount)
	if err != nil || reqAmount.Value.Cmp(new(big.Int)) <= 0 {
		res.Error = "Invalid amount"
		return
	}

	// parse callback
	if req.Callback != "" {
		u, err := url.Parse(req.Callback)
		if err != nil || !u.IsAbs() || !(u.Scheme == "http" || u.Scheme == "https") {
			res.Error = "Invalid callback"
			return
		}
	}

	// enqueue
	if dups, ok := h.api.EnqueueSendingHTTP(req.ID, req.Service, req.Callback, reqAddr, reqAmount, reqToken); !ok {
		if dups {
			res.Error = "Request with the same ID registered"
		} else {
			res.Error = "Internal failure"
			res.Status = gohttp.StatusInternalServerError
		}
		return
	}

	// success
	res.Success = true
	res.Error = ""
	res.Status = gohttp.StatusOK
}
