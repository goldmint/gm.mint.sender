package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/amount"
)

// PublishSentEvent sends a sending completion notification
func (h *HTTP) PublishSentEvent(
	success bool,
	msgerr string,
	service, requestID, callbackURL string,
	to mint.PublicKey,
	token mint.Token,
	amo *amount.Amount,
	digest *mint.Digest,
) error {
	// metrics
	if h.metrics != nil {
		defer func(t time.Time) {
			h.metrics.NotificationDuration.Observe(time.Since(t).Seconds())
		}(time.Now())
	}

	transaction := ""
	if digest != nil {
		transaction = (*digest).String()
	}

	event := SentEvent{
		Success:     success,
		Error:       msgerr,
		Service:     service,
		ID:          requestID,
		PublicKey:   to.String(),
		Token:       token.String(),
		Amount:      amo.String(),
		Transaction: transaction,
	}

	b, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// http request
	{
		timeoutSec := 10
		transport := &http.Transport{
			IdleConnTimeout: time.Second * time.Duration(timeoutSec),
		}
		client := &http.Client{
			Timeout:   time.Second * time.Duration(timeoutSec),
			Transport: transport,
		}
		req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(b))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("callback status code is %v", resp.StatusCode)
		}
	}

	return nil
}
