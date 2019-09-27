package http

import (
	"encoding/json"
	"io/ioutil"
	gohttp "net/http"
	"net/url"
	"time"
	"unicode/utf8"

	"github.com/void616/gm-mint-sender/internal/watcher/api/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

// watch processes request to add a wallet to ROI
func (h *HTTP) watch(w gohttp.ResponseWriter, r *gohttp.Request) {
	defer r.Body.Close()

	// metrics
	if h.metrics != nil {
		defer func(t time.Time) {
			h.metrics.RequestDuration.WithLabelValues("watch").Observe(time.Since(t).Seconds())
		}(time.Now())
	}

	// parse
	req := WatchRequest{}
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

	h.logger.WithField("data", len(req.PublicKeys)).Debug("Got watch request")

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

	// unpack base58
	pubs := make([]sumuslib.PublicKey, 0)
	for _, p := range req.PublicKeys {
		pub, err := sumuslib.ParsePublicKey(p)
		if err != nil {
			res.Error = "One or more invalid Base58 public keys"
			return
		}
		pubs = append(pubs, pub)
	}
	if len(pubs) == 0 {
		res.Error = "Empty list of public keys"
		return
	}

	// parse callback
	reqURL, err := url.Parse(req.Callback)
	if err != nil || !reqURL.IsAbs() || !(reqURL.Scheme == "http" || reqURL.Scheme == "https") {
		res.Error = "Invalid callback"
		return
	}

	// add to ROI
	if !h.api.AddWallet(req.Service, types.ServiceHTTP, req.Callback, pubs...) {
		res.Error = "Internal failure"
		res.Status = gohttp.StatusInternalServerError
		return
	}

	// success
	res.Success = true
	res.Error = ""
	res.Status = gohttp.StatusOK
}

// unwatch processes request to remove a wallet from ROI
func (h *HTTP) unwatch(w gohttp.ResponseWriter, r *gohttp.Request) {
	defer r.Body.Close()

	// metrics
	if h.metrics != nil {
		defer func(t time.Time) {
			h.metrics.RequestDuration.WithLabelValues("unwatch").Observe(time.Since(t).Seconds())
		}(time.Now())
	}

	// parse
	req := UnwatchRequest{}
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

	h.logger.WithField("data", len(req.PublicKeys)).Debug("Got unwatch request")

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

	// unpack base58
	pubs := make([]sumuslib.PublicKey, 0)
	for _, p := range req.PublicKeys {
		pub, err := sumuslib.ParsePublicKey(p)
		if err != nil {
			res.Error = "One or more invalid Base58 public keys"
			return
		}
		pubs = append(pubs, pub)
	}
	if len(pubs) == 0 {
		res.Error = "Empty list of public keys"
		return
	}

	// add to ROI
	if !h.api.RemoveWallet(req.Service, pubs...) {
		res.Error = "Internal failure"
		res.Status = gohttp.StatusInternalServerError
		return
	}

	// success
	res.Success = true
	res.Error = ""
	res.Status = gohttp.StatusOK
}
